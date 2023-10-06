package service

import (
	"database/sql"
	"effective_mobile/entities"
	"effective_mobile/repositories"
	"effective_mobile/repositories/impl"
)

type PersonService struct {
	PersonRepository repositories.PersonRepository
}

func NewPersonService(db *sql.DB) *PersonService {
	personRepository := impl.NewPersonRepository(db)
	return &PersonService{
		PersonRepository: personRepository,
	}
}

func (s *PersonService) CreatePerson(person *entities.Person) (*entities.Person, error) {
	return s.PersonRepository.CreatePerson(person)
}

func (s *PersonService) GetPersonByID(personID int) (*entities.Person, error) {
	return s.PersonRepository.GetPersonByID(personID)
}

func (s *PersonService) GetPersonByName(name string) (*entities.Person, error) {
	return s.PersonRepository.GetPersonByName(name)
}

func (s *PersonService) UpdatePerson(person *entities.Person) (*entities.Person, error) {
	return s.PersonRepository.UpdatePerson(person)
}

func (s *PersonService) DeletePerson(personId int) bool {
	return s.PersonRepository.DeletePerson(personId)
}
