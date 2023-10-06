package test

import (
	"effective_mobile/entities"
	"effective_mobile/service"
	"testing"
)

type MockPersonRepository struct {
	createPersonFunc    func(person *entities.Person) (*entities.Person, error)
	getPersonByIDFunc   func(personID int) (*entities.Person, error)
	getPersonByNameFunc func(name string) (*entities.Person, error)
	updatePersonFunc    func(person *entities.Person) (*entities.Person, error)
	deletePersonFunc    func(personID int) bool
}

func (m *MockPersonRepository) CreatePerson(person *entities.Person) (*entities.Person, error) {
	return m.createPersonFunc(person)
}

func (m *MockPersonRepository) GetPersonByID(personID int) (*entities.Person, error) {
	return m.getPersonByIDFunc(personID)
}

func (m *MockPersonRepository) GetPersonByName(name string) (*entities.Person, error) {
	return m.getPersonByNameFunc(name)
}

func (m *MockPersonRepository) UpdatePerson(person *entities.Person) (*entities.Person, error) {
	return m.updatePersonFunc(person)
}

func (m *MockPersonRepository) DeletePerson(personID int) bool {
	return m.deletePersonFunc(personID)
}

func TestPersonService_CreatePerson(t *testing.T) {
	mockRepo := &MockPersonRepository{
		createPersonFunc: func(person *entities.Person) (*entities.Person, error) {
			return &entities.Person{ID: 1, Name: "John"}, nil
		},
	}

	service := &service.PersonService{PersonRepository: mockRepo}

	personToCreate := &entities.Person{Name: "John"}
	createdPerson, err := service.CreatePerson(personToCreate)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if createdPerson.Name != "John" {
		t.Errorf("Expected created person's name to be 'John', got '%s'", createdPerson.Name)
	}
}

func TestPersonService_GetPersonByID(t *testing.T) {
	mockRepo := &MockPersonRepository{
		getPersonByIDFunc: func(personID int) (*entities.Person, error) {
			return &entities.Person{ID: 1, Name: "John"}, nil
		},
	}

	service := &service.PersonService{PersonRepository: mockRepo}

	personIDToGet := 1
	person, err := service.GetPersonByID(personIDToGet)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if person.Name != "John" {
		t.Errorf("Expected person's name to be 'John', got '%s'", person.Name)
	}
}

func TestPersonService_UpdatePerson(t *testing.T) {
	mockRepo := &MockPersonRepository{
		updatePersonFunc: func(person *entities.Person) (*entities.Person, error) {
			return person, nil
		},
	}

	service := &service.PersonService{PersonRepository: mockRepo}

	personToUpdate := &entities.Person{ID: 1, Name: "UpdatedJohn"}
	updatedPerson, err := service.UpdatePerson(personToUpdate)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if updatedPerson.Name != "UpdatedJohn" {
		t.Errorf("Expected updated person's name to be 'UpdatedJohn', got '%s'", updatedPerson.Name)
	}
}

func TestPersonService_DeletePerson(t *testing.T) {
	mockRepo := &MockPersonRepository{
		deletePersonFunc: func(personID int) bool {
			return true
		},
	}

	service := &service.PersonService{PersonRepository: mockRepo}

	personIDToDelete := 1
	success := service.DeletePerson(personIDToDelete)

	if !success {
		t.Errorf("Expected successful deletion, but deletePersonFunc returned false")
	}
}
