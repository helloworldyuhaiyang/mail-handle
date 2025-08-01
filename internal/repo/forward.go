package repo

type ForwardTargetRepo interface {
	FindEmailByName(targetName string) (email string, err error)
}
