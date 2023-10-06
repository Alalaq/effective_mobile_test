package entities

type Person struct {
	ID          int    `db:"id"`
	Name        string `db:"name"`
	Surname     string `db:"surname"`
	Patronymic  string `db:"patronymic"`
	Age         int    `db:"age"`
	Gender      string `db:"gender"`
	Nationality string `db:"nationality"`
}

func NewPerson(id, age int, name, surname, patronymic, gender, nationality string) *Person {
	return &Person{
		ID:          id,
		Name:        name,
		Surname:     surname,
		Patronymic:  patronymic,
		Age:         age,
		Gender:      gender,
		Nationality: nationality,
	}
}
