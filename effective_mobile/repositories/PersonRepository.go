package repositories

import (
	"effective_mobile/entities"
)

type PersonRepository interface {
	CreatePerson(person *entities.Person) (*entities.Person, error)
	GetPersonByID(personID int) (*entities.Person, error)
	GetPersonByName(name string) (*entities.Person, error)
	UpdatePerson(person *entities.Person) (*entities.Person, error)
	DeletePerson(personID int) bool
}
