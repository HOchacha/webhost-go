package hosting_service

type Service interface {
	CreateHosting(userID int64, username string) (*Hosting, error)
	//StartHosting(id int64) error
	//StopHosting(id int64) error
	//RestartHosting(id int64) error
	//DeleteHosting(id int64) error

	//ListHostingForUser(userID int64) ([]*Hosting, error)
	//GetHostingDetail(id int64) (*Hosting, error)

	//UpdatePlan(id int64, newPlan string) error // optional
}
