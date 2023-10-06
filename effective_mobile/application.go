package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/graphql-go/graphql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"

	"effective_mobile/entities"
	"effective_mobile/service"
)

var producerConfig = sarama.NewConfig()
var redisClient *redis.Client
var personType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Person",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.Int,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"surname": &graphql.Field{
			Type: graphql.String,
		},
		"patronymic": &graphql.Field{
			Type: graphql.String,
		},
		"age": &graphql.Field{
			Type: graphql.Int,
		},
		"gender": &graphql.Field{
			Type: graphql.String,
		},
		"nationality": &graphql.Field{
			Type: graphql.String,
		},
	},
})

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Retrieve environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	kafkaBroker := os.Getenv("KAFKA_BROKER")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	port := os.Getenv("PORT")

	db, err := sql.Open("postgres", "postgres://"+dbUser+":"+dbPassword+"@"+dbHost+":"+dbPort+"/"+dbName+"?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	personService := service.NewPersonService(db)

	var queryType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"person": &graphql.Field{
				Type: personType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(int)
					person, err := personService.GetPersonByID(id)
					return person, err
				},
			},
		},
	})

	var mutationType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createPerson": &graphql.Field{
				Type: personType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"surname": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"patronymic": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name, _ := p.Args["name"].(string)
					surname, _ := p.Args["surname"].(string)
					patronymic, _ := p.Args["patronymic"].(string)
					newPerson := &entities.Person{
						Name:       name,
						Surname:    surname,
						Patronymic: patronymic,
					}
					createdPerson, err := personService.CreatePerson(newPerson)
					return createdPerson, err
				},
			},
			"updatePerson": &graphql.Field{
				Type: personType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"surname": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"patronymic": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(int)
					name, _ := p.Args["name"].(string)
					surname, _ := p.Args["surname"].(string)
					patronymic, _ := p.Args["patronymic"].(string)
					updatedPerson := &entities.Person{
						ID:         id,
						Name:       name,
						Surname:    surname,
						Patronymic: patronymic,
					}
					updatedPerson, err := personService.UpdatePerson(updatedPerson)
					return updatedPerson, err
				},
			},
			"deletePerson": &graphql.Field{
				Type: graphql.Boolean,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, _ := p.Args["id"].(int)
					success := personService.DeletePerson(id)
					return success, nil
				},
			},
		},
	})

	var schema, _ = graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType, // Add mutation type
	})

	broker := kafkaBroker
	topic := kafkaTopic

	consumerConfig := sarama.NewConfig()
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer([]string{broker}, consumerConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		for message := range partitionConsumer.Messages() {
			// Process Kafka message
			var inputPerson entities.Person
			if err := json.Unmarshal(message.Value, &inputPerson); err != nil {
				fmt.Printf("Error processing message: %v\n", err)
				sendToFailedQueue(message.Value)
				continue
			}

			enrichedPerson, err := enrichPerson(&inputPerson)
			if err != nil {
				fmt.Printf("Error enriching person data: %v\n", err)
				continue
			}

			createdPerson, err := personService.CreatePerson(enrichedPerson)
			if err != nil {
				fmt.Printf("Error creating person: %v\n", err)
			} else {
				fmt.Printf("Created Person: %+v\n", createdPerson)
			}
		}
	}()

	router := gin.Default()

	router.POST("/api/people", func(c *gin.Context) {
		var inputPerson entities.Person
		if err := c.BindJSON(&inputPerson); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		enrichedPerson, err := enrichPerson(&inputPerson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error enriching person data"})
			return
		}

		createdPerson, err := personService.CreatePerson(enrichedPerson)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating person"})
			return
		}

		c.JSON(http.StatusCreated, createdPerson)
	})

	router.GET("/api/people/:id", func(c *gin.Context) {
		personID := c.Param("id")

		personIDInt, err := strconv.Atoi(personID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID"})
			return
		}

		person, err := personService.GetPersonByID(personIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching person"})
			return
		}

		if person == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		c.JSON(http.StatusOK, person)
	})

	router.PUT("/api/people/:id", func(c *gin.Context) {
		// Parse the person ID from the request URL
		personID := c.Param("id")

		personIDInt, err := strconv.Atoi(personID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID"})
			return
		}

		var updatedPersonData entities.Person
		if err := c.BindJSON(&updatedPersonData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		existingPerson, err := personService.GetPersonByID(personIDInt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching person"})
			return
		}

		if existingPerson == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		updatedPerson := &entities.Person{
			ID:          existingPerson.ID,
			Name:        updatedPersonData.Name,
			Surname:     updatedPersonData.Surname,
			Patronymic:  updatedPersonData.Patronymic,
			Age:         updatedPersonData.Age,
			Gender:      updatedPersonData.Gender,
			Nationality: updatedPersonData.Nationality,
		}

		updatedPerson, updateErr := personService.UpdatePerson(updatedPerson)
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating person"})
			return
		}

		c.JSON(http.StatusOK, updatedPerson)
	})

	router.DELETE("/api/people/:id", func(c *gin.Context) {
		// Parse the person ID from the request URL
		personID := c.Param("id")

		personIDInt, err := strconv.Atoi(personID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid person ID"})
			return
		}

		success := personService.DeletePerson(personIDInt)
		if !success {
			c.JSON(http.StatusNotFound, gin.H{"error": "Person not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Person deleted successfully"})
	})

	router.POST("/graphql", func(c *gin.Context) {
		var requestBody map[string]interface{}
		if err := c.BindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: requestBody["query"].(string),
		})

		if len(result.Errors) > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"errors": result.Errors})
			return
		}

		c.JSON(http.StatusOK, result.Data)
	})

	err = router.Run(fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func init() {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})
}

func sendToFailedQueue(message []byte) {
	producer, err := sarama.NewSyncProducer([]string{"localhost:9092"}, producerConfig)
	if err != nil {
		fmt.Printf("Error creating Kafka producer: %v\n", err)
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			fmt.Printf("Error closing Kafka producer: %v\n", err)
		}
	}()

	kafkaMessage := &sarama.ProducerMessage{
		Topic: "FIO_FAILED",
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(kafkaMessage)
	if err != nil {
		fmt.Printf("Error sending message to FIO_FAILED Kafka queue: %v\n", err)
		return
	}

	fmt.Printf("Sent message to FIO_FAILED Kafka queue - Partition: %d, Offset: %d\n", partition, offset)
}

func enrichPerson(inputPerson *entities.Person) (*entities.Person, error) {
	age, err := fetchAge(inputPerson.Name)
	if err != nil {
		return nil, err
	}
	inputPerson.Age = age

	gender, err := fetchGender(inputPerson.Name)
	if err != nil {
		return nil, err
	}
	inputPerson.Gender = gender

	nationality, err := fetchNationality(inputPerson.Name)
	if err != nil {
		return nil, err
	}
	inputPerson.Nationality = nationality

	return inputPerson, nil
}

func fetchAge(name string) (int, error) {
	age, err := redisClient.Get(context.Background(), "age:"+name).Int()
	if err == nil {
		return age, nil
	}

	url := fmt.Sprintf("https://api.agify.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Failed to fetch age data: %s", resp.Status)
	}

	var ageData map[string]int
	json.NewDecoder(resp.Body).Decode(&ageData)

	age = ageData["age"]
	if age == 0 {
		return 0, fmt.Errorf("Age data not found")
	}

	err = redisClient.Set(context.Background(), "age:"+name, age, 0).Err()
	if err != nil {
		fmt.Printf("Failed to cache age data in Redis: %v\n", err)
	}

	return age, nil
}
func fetchGender(name string) (string, error) {
	gender, err := redisClient.Get(context.Background(), "gender:"+name).Result()
	if err == nil {
		return gender, nil
	}

	url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to fetch gender data: %s", resp.Status)
	}

	var genderData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&genderData)
	if err != nil {
		return "", err
	}

	gender, ok := genderData["gender"].(string)
	if !ok {
		return "", fmt.Errorf("Gender data not found")
	}

	err = redisClient.Set(context.Background(), "gender:"+name, gender, 0).Err()
	if err != nil {
		fmt.Printf("Failed to cache gender data in Redis: %v\n", err)
	}

	return gender, nil
}

func fetchNationality(name string) (string, error) {
	nationality, err := redisClient.Get(context.Background(), "nationality:"+name).Result()
	if err == nil {
		return nationality, nil
	}

	url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Failed to fetch nationality data: %s", resp.Status)
	}

	var nationalityData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&nationalityData)
	if err != nil {
		return "", err
	}

	countryList, ok := nationalityData["country"].([]interface{})
	if !ok || len(countryList) == 0 {
		return "", fmt.Errorf("Nationality data not found")
	}

	firstCountry, ok := countryList[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("Nationality data not found")
	}

	countryCode, ok := firstCountry["country_id"].(string)
	if !ok {
		return "", fmt.Errorf("Nationality data not found")
	}

	err = redisClient.Set(context.Background(), "nationality:"+name, countryCode, 0).Err()
	if err != nil {
		fmt.Printf("Failed to cache nationality data in Redis: %v\n", err)
	}

	return countryCode, nil
}
