package impl

import (
	"database/sql"
	"effective_mobile/entities"
	"errors"
)

type PersonRepositoryImpl struct {
	db *sql.DB
}

func NewPersonRepository(db *sql.DB) *PersonRepositoryImpl {
	return &PersonRepositoryImpl{db: db}
}

func (r *PersonRepositoryImpl) CreatePerson(person *entities.Person) (*entities.Person, error) {
	// Insert the person without the RETURNING clause
	insertQuery := `
		INSERT INTO persons (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(insertQuery, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality)
	if err != nil {
		return nil, err
	}

	var id int
	err = r.db.QueryRow("SELECT LASTVAL()").Scan(&id)
	if err != nil {
		return nil, err
	}

	person.ID = id
	return person, nil
}

func (r *PersonRepositoryImpl) GetPersonByID(personID int) (*entities.Person, error) {
	query := "SELECT id, name, surname, patronymic, age, gender, nationality FROM persons WHERE id = $1"

	var person entities.Person
	err := r.db.QueryRow(query, personID).Scan(&person.ID, &person.Name, &person.Surname, &person.Patronymic, &person.Age, &person.Gender, &person.Nationality)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Person not found
		}
		return nil, err
	}

	return &person, nil
}

func (r *PersonRepositoryImpl) GetPersonByName(name string) (*entities.Person, error) {
	query := "SELECT id, name, surname, patronymic, age, gender, nationality FROM persons WHERE name = $1"

	var person entities.Person
	err := r.db.QueryRow(query, name).Scan(&person.ID, &person.Name, &person.Surname, &person.Patronymic, &person.Age, &person.Gender, &person.Nationality)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Person not found
		}
		return nil, err
	}

	return &person, nil
}

func (r *PersonRepositoryImpl) UpdatePerson(person *entities.Person) (*entities.Person, error) {
	query := `
		UPDATE persons
		SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5, nationality = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(query, person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality, person.ID)
	if err != nil {
		return nil, err
	}

	return person, nil
}

func (r *PersonRepositoryImpl) DeletePerson(personID int) bool {
	query := "DELETE FROM persons WHERE id = $1"

	result, err := r.db.Exec(query, personID)
	if err != nil {
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false
	}

	return rowsAffected > 0
}
