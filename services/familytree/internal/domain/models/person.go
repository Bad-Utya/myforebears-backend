package models

import "github.com/google/uuid"

type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
)

type Person struct {
	ID            uuid.UUID
	TreeID        uuid.UUID
	FirstName     string
	LastName      string
	Patronymic    string
	Gender        Gender
	AvatarPhotoID *uuid.UUID
}
