package reddit

type set map[string]struct{}

func (s set) Add(v string) {
	s[v] = struct{}{}
}

func (s set) Delete(v string) {
	delete(s, v)
}

func (s set) Len() int {
	return len(s)
}

func (s set) Exists(v string) bool {
	_, ok := s[v]
	return ok
}
