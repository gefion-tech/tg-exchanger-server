package db

/*
	Родительский интерфейс для работы с БД
*/
type SQLStoreI interface {
	/*
		Интерфейс для взаимодействия с объектом пользователя и
		его связанными с ним дочерними репозиториями в БД
	*/
	User() UserRepository

	AdminPanel() AdminPanelRepository
}
