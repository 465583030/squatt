package topic

// Topic structure
type Topic struct {
	name string
}

// Name of the topic
func (t *Topic) Name() string {
	return t.name
}

// NewTopic returns a new topic with the given name
func NewTopic(name string) *Topic {
	return &Topic{
		name: name,
	}
}
