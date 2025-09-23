package database

func NewDatabaseService(config Config) DatabaseService {
	return NewManager(config)
}
