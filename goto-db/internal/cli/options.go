package cli

type Options struct {
	DBName       string
	Environment  string
	Engine       string
	JenkinsAgent string
	Refresh      bool
	LocalPort    int
	User         string
	DBURL        string
}
