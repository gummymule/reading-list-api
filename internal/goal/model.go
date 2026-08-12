package goal

type ReadingGoal struct {
	Year    int `json:"year"`
	Target  int `json:"target"`
	Current int `json:"current"`
}

type UpdateGoalInput struct {
	Year   int `json:"year"`
	Target int `json:"target"`
}
