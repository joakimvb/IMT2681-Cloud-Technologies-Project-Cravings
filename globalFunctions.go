package cravings

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// DoRequest sends a new http request
func DoRequest(url string, c *http.Client) (*http.Response, error) {
	var resp *http.Response

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return resp, errors.Wrap(err, "Unable to request: "+url+err.Error())
	}

	resp, err = c.Do(req)
	if err != nil {
		return resp, errors.Wrap(err, "Unable to get: "+url+err.Error())
	}

	return resp, err
}

// QueryGet func to read query with name s, sets to def value if not defined
func QueryGet(s string, def string, r *http.Request) string {
	query := r.URL.Query().Get(s) // gets app key or app id
	if query == "" {              //if it is empty
		query = def //set to default value
	}

	return query
}

// CallURL post webhooks to webhooks.site
func CallURL(event string, s interface{}, w http.ResponseWriter) error {
	webhooks, err := DBReadAllWebhooks(w) // gets all webhooks
	if err != nil {
		fmt.Fprintln(w, "Could not retrieve documents from webhooks collection: "+err.Error(), http.StatusInternalServerError)
		return err
	}

	for i := range webhooks { // loops through all webhooks
		if webhooks[i].Event == event { // see if webhooks.event is same as event
			var request = s

			requestBody, err := json.Marshal(request)
			if err != nil {
				fmt.Fprintln(w, "Can not encode: "+err.Error(), http.StatusInternalServerError)
				return err
			}

			fmt.Println("Attempting invocation of URL " + webhooks[i].URL + "...")

			resp, err := http.Post(webhooks[i].URL, "json", bytes.NewReader(requestBody)) // post webhook to webhooks.site
			if err != nil {
				fmt.Fprintln(w, "Error in HTTP request: "+err.Error(), http.StatusBadRequest)
				return err
			}

			defer resp.Body.Close() // close body
		}
	}

	return nil
}

// ReadIngredients splits up the ingredient name from the quantity from the URL
func ReadIngredients(ingredients []string, w http.ResponseWriter) ([]Ingredient, error) {
	IngredientList := []Ingredient{}
	defVal := 1.0 //default value for quantity if not set

	var err error

	for i := range ingredients {
		ingredient := strings.Split(ingredients[i], "|") //splits up the string 'name|quantity|unit'
		ingredientTemp := Ingredient{}

		if len(ingredient) < 3 { //if quantity value is set
			return IngredientList, errors.New(
				"Failed to read ingredient list. ?ingredients={name|quantity|unit}_{...}")
		}

		ingredientTemp.Name = ingredient[0] //name of the ingredient

		allowed := false

		for _, unit := range AllowedUnit {
			if ingredient[2] == unit {
				allowed = true
			}
		}

		if !allowed {
			return IngredientList, errors.New(ingredient[2] + " is not an allowed unit.")
		}

		ingredientTemp.Unit = ingredient[2]                                  //sets the unit
		ingredientTemp.Quantity, err = strconv.ParseFloat(ingredient[1], 64) //sets the quantity

		if err != nil { //if error: set quantity to defVal
			ingredientTemp.Quantity = defVal
		}

		IngredientList = append(IngredientList, ingredientTemp)
	}

	return IngredientList, nil
}

// CalcRemaining calculates the nutritional value from one ingredient to another.
// If subtract is true, it also subtracts quantity from rec in ing
func CalcRemaining(ing Ingredient, rec Ingredient, subtract bool) Ingredient {
	if ing.Unit != rec.Unit { //if the ingredients measures in different units
		if strings.Contains(rec.Unit, "spoon") { //if rec contains spoon unit
			noOfSpoons := ing.Calories / (rec.Calories / rec.Quantity) //calculates number of spoons for ing
			unitPerSpoon := ing.Quantity / noOfSpoons                  //how many calories in one spoon
			rec.Quantity *= unitPerSpoon                               //spoons times with calories per
			rec.Unit = ing.Unit                                        //spoon to get the same unit
		} else {
			ConvertUnit(&ing, rec.Unit) //convert ing to same unit as rec
		}
	}

	if subtract {
		ing.Quantity -= rec.Quantity
	}

	ing.Nutrients = rec.Nutrients //sets all the labels and units for nutrients

	//calculates the values for 1 ingredient, then multiplies by ingredients quantity
	ing.Calories = (rec.Calories / rec.Quantity) * ing.Quantity
	ing.Weight = (rec.Weight / rec.Quantity) * ing.Quantity
	ing.Nutrients.Carbohydrate.Quantity = (rec.Nutrients.Carbohydrate.Quantity / rec.Quantity) * ing.Quantity
	ing.Nutrients.Energy.Quantity = (rec.Nutrients.Energy.Quantity / rec.Quantity) * ing.Quantity
	ing.Nutrients.Fat.Quantity = (rec.Nutrients.Fat.Quantity / rec.Quantity) * ing.Quantity
	ing.Nutrients.Protein.Quantity = (rec.Nutrients.Protein.Quantity / rec.Quantity) * ing.Quantity
	ing.Nutrients.Sugar.Quantity = (rec.Nutrients.Sugar.Quantity / rec.Quantity) * ing.Quantity

	return ing
}

// CalcNutrition calculates nutritional info for given ingredient
func CalcNutrition(ing Ingredient, w http.ResponseWriter) (Ingredient, error) {
	temping, err := DBReadIngredientByName(ing.Name, w) //gets the ingredient with the same name from firebase
	if err != nil {
		return ing, errors.Wrap(err, "Could not read ingredient by name "+err.Error())
	}

	ing.ID = temping.ID               // add ID to ing since it's a copy
	ing.Nutrients = temping.Nutrients // reset nutrients to nutrients for 1g or 1l

	switch ing.Unit {
	case "kg":
		ConvertUnit(&ing, "g")
	case "g":
		ConvertUnit(&ing, "g")
	case "l":
		ConvertUnit(&ing, "l")
	case "dl":
		ConvertUnit(&ing, "l")
	case "cl":
		ConvertUnit(&ing, "l")
	case "ml":
		ConvertUnit(&ing, "l")
	case "pc":
	case "tablespoon":
		err := GetNutrients(&ing, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	case "teaspoon":
		err := GetNutrients(&ing, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	ing.Calories = temping.Calories * ing.Quantity //calculates calories based on ingredients quantity
	ing.Weight = temping.Weight * ing.Quantity     //calculates weight based on ingredients quantity

	// Calc nutrition :
	ing.Nutrients.Carbohydrate.Quantity *= ing.Quantity
	ing.Nutrients.Energy.Quantity *= ing.Quantity
	ing.Nutrients.Fat.Quantity *= ing.Quantity
	ing.Nutrients.Protein.Quantity *= ing.Quantity
	ing.Nutrients.Sugar.Quantity *= ing.Quantity

	return ing, nil
}

// ConvertUnit converts units for ingredients, and changes their quantity respectively.
func ConvertUnit(ing *Ingredient, unitConvertTo string) {
	if ing.Unit == "kg" && unitConvertTo == "g" {
		ing.Quantity *= 1000
		ing.Unit = unitConvertTo
	}

	if ing.Unit == "g" && unitConvertTo == "kg" {
		ing.Quantity /= 1000
		ing.Unit = unitConvertTo
	}

	if unitConvertTo == "l" {
		switch ing.Unit {
		case "dl":
			ing.Quantity /= 10
		case "cl":
			ing.Quantity /= 100
		case "ml":
			ing.Quantity /= 1000
		}

		ing.Unit = unitConvertTo
	}

	if unitConvertTo == "dl" {
		switch ing.Unit {
		case "l":
			ing.Quantity *= 10
		case "cl":
			ing.Quantity /= 10
		case "ml":
			ing.Quantity /= 100
		}

		ing.Unit = unitConvertTo
	}

	if unitConvertTo == "cl" {
		switch ing.Unit {
		case "dl":
			ing.Quantity *= 10
		case "l":
			ing.Quantity *= 100
		case "ml":
			ing.Quantity /= 10
		}

		ing.Unit = unitConvertTo
	}

	if unitConvertTo == "ml" {
		switch ing.Unit {
		case "cl":
			ing.Quantity *= 10
		case "dl":
			ing.Quantity *= 100
		case "l":
			ing.Quantity *= 1000
		}

		ing.Unit = unitConvertTo
	}
}

// InitAPICredentials func opens up local file and reads the application id and key from that file
func InitAPICredentials() error {
	//  Opens local file which contains application id and key
	file, err := os.Open("appIdAndKey.txt")
	if err != nil {
		fmt.Println("Error: Unable to open file " + err.Error())
		return err
	}
	defer file.Close()
	//  Scans the lines of the file
	scanner := bufio.NewScanner(file)
	scanner.Scan()
	AppID = scanner.Text()
	scanner.Scan()
	AppKey = scanner.Text()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error: Unable to read the application ID and key from file " + err.Error())
		return err
	}

	return err
}

// UnitCheck func checks the unit measurements of two ingredients and checks if they are of the same type solid/liquid
func UnitCheck(firstIngredient string, secondIngredient string) bool {
	if strings.Contains(firstIngredient, "l") {
		if strings.Contains(secondIngredient, "l") {
			return true
		}
	}

	if strings.Contains(firstIngredient, "g") {
		if strings.Contains(secondIngredient, "g") {
			return true
		}
	}

	if strings.Contains(firstIngredient, "spoon") {
		// table/teaspoon can be registered as liquid or solid
		return true
	}

	if strings.Contains(firstIngredient, "pc") {
		return true
	}

	return false
}
