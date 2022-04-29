package vo

type DoClusterTaskVO struct {
	TaskCode string
	RecordId int
	Stages []map[int]map[string]string
}