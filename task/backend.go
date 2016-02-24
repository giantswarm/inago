package task

type Backend interface {
	Get(string) (*TaskObject, error)
	Set(*TaskObject) error
}
