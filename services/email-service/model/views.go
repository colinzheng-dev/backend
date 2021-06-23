package model

// TopicInfo is a view of a Topic including only the information
// relevant to other services.
type TopicInfo struct {
	Name        string `json:"name"`
	SendAddress string `json:"send_address"`
}

// Info generates a informational view of a topic from a database
// model.
func (topic *Topic) Info() *TopicInfo {
	return &TopicInfo{
		Name:        topic.Name,
		SendAddress: topic.SendAddress,
	}
}
