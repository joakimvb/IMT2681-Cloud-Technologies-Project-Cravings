package cravings

import (
	"context"

	"cloud.google.com/go/firestore"
)

//  Struct for a recipe containting ingredients used in firebase.go and register.go
type Recipe struct {
	ID          string       `json:"id"`
	Ingredients []Ingredient `json:"ingredients"`
}

//  Struct for an ingredient used in firebase.go and register.go
type Ingredient struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
	Name     string `json:"name"`
	Calories int    `json:"kcal"`
	Weight   int    `json:"weight"`
}

type NutrionalFacts struct {
}

// FirestoreDatabase implements our Database access through Firestore
type FirestoreDatabase struct {
	Ctx    context.Context
	Client *firestore.Client
}
