package reddit

import (
	"context"
	"fmt"
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
//
// Because of the 100 post limit imposed by Reddit when fetching posts, some high-traffic
// streams might drop submissions between API requests, such as when streaming r/all.
func (s *StreamService) Posts(ctx context.Context, subreddit string, opts ...StreamOpt) (<-chan *Post, <-chan error, func()) {
	return doStream(ctx, subreddit, s.getPosts, opts...)
}

func (s *StreamService) getPosts(ctx context.Context, subreddit string) ([]*Post, error) {
	posts, _, err := s.client.Subreddit.NewPosts(ctx, subreddit, &ListOptions{Limit: itemLimit})
	return posts, err
}

// TODO: Generalize these two functions to have the same body... Maybe when generics is released ;)
func (s *StreamService) Actions(ctx context.Context, subreddit string, opts ...StreamOpt) (<-chan *ModAction, <-chan error, func()) {
	return doStream(ctx, subreddit, s.getActions, opts...)
}

func (s *StreamService) getActions(ctx context.Context, subreddit string) ([]*ModAction, error) {
	posts, _, err := s.client.Moderation.Actions(ctx, subreddit, &ListModActionOptions{ListOptions: ListOptions{Limit: itemLimit}})
	return posts, err
}

// TODO: Generalize these two functions to have the same body... Maybe when generics is released ;)
// InboxUnread returns 3 channels, one for comments, DMs, and errors, in that order, plus a function to close the channel
func (s *StreamService) InboxUnread(ctx context.Context, opts ...StreamOpt) (<-chan *Message, <-chan *Message, <-chan error, func()) {
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

		for {
			select {
			case <-ctx.Done():
				errsCh <- ctx.Err()
				return
			case <-ticker.C:
			}
			n++

			messages, err := s.getInboxUnread(ctx)
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

func (s *StreamService) getInboxUnread(ctx context.Context) ([]*Message, error) {
	comments, directMessages, _, err := s.client.Message.InboxUnread(ctx, &ListOptions{Limit: itemLimit})
	return append(comments, directMessages...), err
}

// TODO: Generalize these two functions to have the same body... Maybe when generics is released ;)
// InboxUnread returns 3 channels, one for comments, DMs, and errors, in that order, plus a function to close the channel
func (s *StreamService) Reported(ctx context.Context, subreddit string, opts ...StreamOpt) (<-chan *Post, <-chan *Comment, <-chan error, func()) {

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
	commentsCh := make(chan *Comment)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(postsCh)
			close(commentsCh)
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

		for {
			select {
			case <-ctx.Done():
				errsCh <- ctx.Err()
				return
			case <-ticker.C:
			}
			n++

			posts, comments, err := s.getReported(ctx, subreddit)
			if err != nil {
				errsCh <- err
				if !infinite && n >= streamConfig.MaxRequests {
					break
				}
				continue
			}

			for _, post := range posts {
				id := fmt.Sprintf("%s%d", post.ID, post.NumReports)

				// if this comment id is already part of the set, it means that it and the ones
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

			for _, comment := range comments {
				id := fmt.Sprintf("%s%d", comment.ID, comment.NumReports)

				// if this comment id is already part of the set, it means that it and the ones
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

				commentsCh <- comment
			}

			if !infinite && n >= streamConfig.MaxRequests {
				break
			}
		}
	}()

	return postsCh, commentsCh, errsCh, stop
}

func (s *StreamService) getReported(ctx context.Context, subreddit string) ([]*Post, []*Comment, error) {
	post, comment, _, err := s.client.Moderation.Reported(ctx, subreddit, &ListOptions{Limit: itemLimit})
	return post, comment, err
}

func (s *StreamService) getComments(ctx context.Context, subreddit string) ([]*Comment, error) {
	comments, _, err := s.client.Subreddit.NewComments(ctx, subreddit, &ListOptions{Limit: itemLimit})
	if err != nil {
		return nil, err
	}
	return comments, nil
}

// Comments streams comments from the entirety of reddit, or whatever subreddit is provided
// It returns 2 channels and a function:
//   - a channel into which new comments will be sent
//   - a channel into which any errors will be sent
//   - a function that the client can call once to stop the streaming and close the channels
//
// Because of the 100 post limit imposed by Reddit when fetching comments, some high-traffic
// streams might drop submissions between API requests, such as when streaming r/all.

func (s *StreamService) CommentsStream(ctx context.Context, subreddit string, opts ...StreamOpt) (<-chan *Comment, <-chan error, func()) {
	return doStream(ctx, subreddit, s.getComments, opts...)
}

type HasID interface {
	GetID() string
}

func doStreamWithCustomIDGet[T any](ctx context.Context, subreddit string, getThing func(context.Context, string) ([]T, error), customIDGet func(T) string, opts ...StreamOpt) (<-chan T, <-chan error, func()) {
	streamConfig := &streamConfig{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
		MaxRequests:    0,
	}
	for _, opt := range opts {
		opt(streamConfig)
	}

	ticker := time.NewTicker(streamConfig.Interval)
	commentsCh := make(chan T)
	errsCh := make(chan error)

	var once sync.Once
	stop := func() {
		once.Do(func() {
			ticker.Stop()
			close(commentsCh)
			close(errsCh)
		})
	}

	// originally used the "before" parameter, but if that post gets deleted, subsequent requests
	// would just return empty listings; easier to just keep track of all comment ids encountered
	oldIDs := set{}
	newIDs := set{}

	go func() {
		defer stop()

		infinite := streamConfig.MaxRequests == 0

		var n int
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			n++
			items, err := getThing(ctx, subreddit)
			if err != nil {
				errsCh <- err
				if !infinite && n >= streamConfig.MaxRequests {
					break
				}
				continue
			}

			for _, item := range items {
				id := customIDGet(item)

				// if this item id is already part of the set, it means that it and the ones
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

				commentsCh <- item
			}
			if !infinite && n >= streamConfig.MaxRequests {
				break
			}
		}
	}()

	return commentsCh, errsCh, stop
}

func doStream[T HasID](ctx context.Context, subreddit string, getThing func(context.Context, string) ([]T, error), opts ...StreamOpt) (<-chan T, <-chan error, func()) {
	return doStreamWithCustomIDGet(ctx, subreddit, getThing, func(t T) string {
		return t.GetID()
	}, opts...)
}
