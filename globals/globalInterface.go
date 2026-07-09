package globals

type GlobalRepository interface {
	RetrieveH2hConfigAPIEmail(id int) (*H2hConfig, error)
}
