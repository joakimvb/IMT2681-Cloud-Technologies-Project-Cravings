package cravings

import (
	"time"
)

// StartTime is the timestamp for when the program started
var StartTime time.Time // start run time

// FirestoreCredentials is the credentials file for firestore db
const FirestoreCredentials = "./cloudproject-2a9c2-firebase-adminsdk-0om9b-bca5ed564a.json"

// RecipeCollection is the name of the recipes collection in the database
const RecipeCollection = "recipes"

// IngredientCollection is the name of the ingredients collection in the database
const IngredientCollection = "ingredients"

// TokenCollection is the name of the tokens collection in the database
const TokenCollection = "tokens"

// WebhooksCollection is the name of the webhooks collection in the database
const WebhooksCollection = "webhooks"

// AllowedUnit = list of units of measurement: kilogram, gram, liter, deciliter, mililiter, piece, teaspoon etc.
var AllowedUnit = [8]string{"kg", "g", "l", "dl", "ml", "pc", "tablespoon", "teaspoon"}

// URLRegistration is the url to edamam api for getting nutrition details when registering an ingredient or recipe
var URLRegistration = "https://api.edamam.com/api/nutrition-details"

// AppID is Application ID for external API
var AppID = ""

// AppKey is Key for external API
var AppKey = ""
