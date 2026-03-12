package db

func DefaultLocalPort(engine string) int {
	switch engine {
	case "mysql":
		return 13306
	default:
		return 15432
	}
}

func DefaultPort(engine string) int {
	switch engine {
	case "mysql":
		return 3306
	default:
		return 5432
	}
}

func DefaultClient(engine string) string {
	switch engine {
	case "mysql":
		return "mysql"
	default:
		return "psql"
	}
}
