package reddit

import (
	"context"
	"sync"
	"time"
)

// StreamService allows streaming new content from Reddit as it appears.
type StreamService struct {
	client *Client
}

var itemLimit = 100

// Posts streams posts from the specified subreddit.
// It returns 2 channels and a function:
//   - a channel into which new posts will be sent
//   - a channel into which any errors will be sent
//   - a function that the client can call once to stop the streaming and close the channels
// Because of the 100 post limit imposed by Reddit when fetching posts, some high-traffic
// streams might drop submissions between API requests, such as when streaming r/all.
func (s *StreamService) Posts(subreddit string, opts ...StreamOpt) (<-chan *Post, <-chan error, func()) {
	streamConfig := &streamConfig{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
		MaxRequests:    0,
	}
	for _, opt := range opts {
		opt(streamConfig)
	}

	ticker := time.NewTicker(streamConfig.Interval)
	postsCh := make(chan *Post)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(postsCh)
			close(errsCh)
		})
	}

	// originally used the "before" parameter, but if that post gets deleted, subsequent requests
	// would just return empty listings; easier to just keep track of all post ids encountered
	oldIDs := set{}
	newIDs := set{}

	go func() {
		defer stop()

		var n int
		infinite := streamConfig.MaxRequests == 0

		for ; ; <-ticker.C {
			n++

			posts, err := s.getPosts(subreddit)
			if err != nil {
				errsCh <- err
				if !infinite && n >= streamConfig.MaxRequests {
					break
				}
				continue
			}

			for _, post := range posts {
				id := post.FullID

				// if this post id is already part of the set, it means that it and the ones
				// after it in the list have already been streamed, so break out of the loop
				if newIDs.Exists(id) || oldIDs.Exists(id) {
					break
				}
				newIDs.Add(id)

				// If the new map is 10 times larger than item limit, make it the old map and clear it
				if len(newIDs) >= itemLimit*10 {
					oldIDs = newIDs
					newIDs = make(map[string]struct{})
				}

				if streamConfig.DiscardInitial {
					streamConfig.DiscardInitial = false
					break
				}

				postsCh <- post
			}

			if !infinite && n >= streamConfig.MaxRequests {
				break
			}
		}
	}()

	return postsCh, errsCh, stop
}

func (s *StreamService) getPosts(subreddit string) ([]*Post, error) {
	posts, _, err := s.client.Subreddit.NewPosts(context.Background(), subreddit, &ListOptions{Limit: itemLimit})
	return posts, err
}

// TODO: Generalize these two functions to have the same body... Maybe when generics is released ;)
func (s *StreamService) Actions(subreddit string, opts ...StreamOpt) (<-chan *ModAction, <-chan error, func()) {
	streamConfig := &streamConfig{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
		MaxRequests:    0,
	}
	for _, opt := range opts {
		opt(streamConfig)
	}

	ticker := time.NewTicker(streamConfig.Interval)
	postsCh := make(chan *ModAction)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(postsCh)
			close(errsCh)
		})
	}

	// originally used the "before" parameter, but if that post gets deleted, subsequent requests
	// would just return empty listings; easier to just keep track of all post ids encountered
	oldIDs := set{}
	newIDs := set{}

	go func() {
		defer stop()

		var n int
		infinite := streamConfig.MaxRequests == 0

		for ; ; <-ticker.C {
			n++

			posts, err := s.getActions(subreddit)
			if err != nil {
				errsCh <- err
				if !infinite && n >= streamConfig.MaxRequests {
					break
				}
				continue
			}

			for _, post := range posts {
				id := post.ID

				// if this post id is already part of the set, it means that it and the ones
				// after it in the list have already been streamed, so break out of the loop
				if newIDs.Exists(id) || oldIDs.Exists(id) {
					break
				}
				newIDs.Add(id)

				// If the new map is 10 times larger than item limit, make it the old map and clear it
				if len(newIDs) >= itemLimit*10 {
					oldIDs = newIDs
					newIDs = make(map[string]struct{})
				}

				if streamConfig.DiscardInitial {
					streamConfig.DiscardInitial = false
					break
				}

				postsCh <- post
			}

			if !infinite && n >= streamConfig.MaxRequests {
				break
			}
		}
	}()

	return postsCh, errsCh, stop
}

func (s *StreamService) getActions(subreddit string) ([]*ModAction, error) {
	posts, _, err := s.client.Moderation.Actions(context.Background(), subreddit, &ListModActionOptions{ListOptions: ListOptions{Limit: itemLimit}})
	return posts, err
}

// TODO: Generalize these two functions to have the same body... Maybe when generics is released ;)
// InboxUnread returns 3 channels, one for comments, DMs, and errors, in that order, plus a function to close the channel
func (s *StreamService) InboxUnread(opts ...StreamOpt) (<-chan *Message, <-chan *Message, <-chan error, func()) {
	streamConfig := &streamConfig{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
		MaxRequests:    0,
	}
	for _, opt := range opts {
		opt(streamConfig)
	}

	ticker := time.NewTicker(streamConfig.Interval)
	commentsCh := make(chan *Message)
	dmsCh := make(chan *Message)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(commentsCh)
			close(dmsCh)
			close(errsCh)
		})
	}

	// originally used the "before" parameter, but if that post gets deleted, subsequent requests
	// would just return empty listings; easier to just keep track of all post ids encountered
	oldIDs := set{}
	newIDs := set{}

	go func() {
		defer stop()

		var n int
		infinite := streamConfig.MaxRequests == 0

		for ; ; <-ticker.C {
			n++

			messages, err := s.getInboxUnread()
			if err != nil {
				errsCh <- err
				if !infinite && n >= streamConfig.MaxRequests {
					break
				}
				continue
			}

			for _, post := range messages {
				id := post.ID

				// if this post id is already part of the set, it means that it and the ones
				// after it in the list have already been streamed, so break out of the loop
				if newIDs.Exists(id) || oldIDs.Exists(id) {
					break
				}
				newIDs.Add(id)

				// If the new map is 10 times larger than item limit, make it the old map and clear it
				if len(newIDs) >= itemLimit*10 {
					oldIDs = newIDs
					newIDs = make(map[string]struct{})
				}

				if streamConfig.DiscardInitial {
					streamConfig.DiscardInitial = false
					break
				}

				if post.IsComment {
					commentsCh <- post
				} else {
					dmsCh <- post
				}
			}

			if !infinite && n >= streamConfig.MaxRequests {
				break
			}
		}
	}()

	return commentsCh, dmsCh, errsCh, stop
}

func (s *StreamService) getInboxUnread() ([]*Message, error) {
	comments, directMessages, _, err := s.client.Message.InboxUnread(context.Background(), &ListOptions{Limit: itemLimit})
	return append(comments, directMessages...), err
}
