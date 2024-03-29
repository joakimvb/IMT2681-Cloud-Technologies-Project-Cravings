package cravings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const caseing = "ingredient"
const caserec = "recipe"

// HandlerFood which registers or view either an ingredient or a recipe
// Whenever calling this endpoint in the browser, it is only possible to view the food,
// to register food, one has to post the .json body
func HandlerFood(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json") // JSON http header

	parts := strings.Split(r.URL.Path, "/")

	endpoint := parts[3] // Store the query which represents either recipe or ingredient
	name := ""

	if len(parts) > 4 {
		name = parts[4]
	}

	if endpoint == "" {
		HandlerNil(w, r)
	}

	switch r.Method {
	case http.MethodGet: // Gets either recipes or ingredients
		switch endpoint {
		case caseing:
			if name != "" { // If ingredient name is specified in URL
				ingr, err := DBReadIngredientByName(name, w) // Get that ingredient

				if err != nil {
					http.Error(w, "Couldn't read ingredient by name: "+err.Error(), http.StatusInternalServerError)
					return
				}

				err = json.NewEncoder(w).Encode(&ingr)

				if err != nil {
					http.Error(w, "Couldn't encode response of ingredient: "+err.Error(), http.StatusBadRequest)
					return
				}
			} else {
				ingredients, err := DBReadAllIngredients(w) // Else retrieve all ingredients
				if err != nil {
					http.Error(w, "Couldn't retrieve ingredients: "+err.Error(), http.StatusBadRequest)
					return
				}

				err = json.NewEncoder(w).Encode(&ingredients)
				if err != nil {
					http.Error(w, "Couldn't encode response: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		case caserec:
			if name != "" { // If user wrote in query for name of recipe
				re := Recipe{}

				re, err := DBReadRecipeByName(name, w) // Get that recipe
				if err != nil {
					http.Error(w, "Couldn't retrieve recipe: "+err.Error(), http.StatusBadRequest)
					return
				}

				err = json.NewEncoder(w).Encode(&re)
				if err != nil {
					http.Error(w, "Couldn't encode response: "+err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				recipes, err := DBReadAllRecipes(w) // Else get all recipes
				if err != nil {
					http.Error(w, "Couldn't retrieve recipes: "+err.Error(), http.StatusBadRequest)
					return
				}

				err = json.NewEncoder(w).Encode(&recipes)
				if err != nil {
					http.Error(w, "Couldn't encode response: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

	// Post either recipes or ingredients to firebase DB
	case http.MethodPost:
		authorised, resp, err := DBCheckAuthorization(w, r) // Check for valid token

		if err != nil {
			http.Error(w, "Authorization failed!\nError: "+err.Error(), http.StatusBadRequest)
			return
		}

		//  To post either one, you have to post it with a POST request with a .json body i.e. Postman
		//  and include the authorization token given by the developers through mail inside the body
		//  Detailed instructions for registering is in the readme
		if authorised {
			switch endpoint {
			case caseing: // Posts ingredient
				RegisterIngredient(w, resp)

			case caserec: // Posts recipe
				RegisterRecipe(w, resp)
			}
		} else if err == nil {
			http.Error(w, "Error: Not authorized! Please use a valid token.", http.StatusUnauthorized)
		}

	case http.MethodDelete:
		authorised, resp, err := DBCheckAuthorization(w, r) // Check for valid token
		if err != nil {
			http.Error(w, "Authorization failed!\nError: "+err.Error(), http.StatusBadRequest)
		}

		if authorised {
			switch endpoint {
			case caseing:
				ing := Ingredient{}

				err := json.Unmarshal(resp, &ing)
				if err != nil {
					http.Error(w, "Could not unmarshal body of request"+err.Error(), http.StatusBadRequest)
					return
				}

				ing, err = DBReadIngredientByName(ing.Name, w) //  Get that ingredient
				if err != nil {
					http.Error(w, "Couldn't retrieve ingredient: "+err.Error(), http.StatusBadRequest)
					return
				}

				insideARecipe, err := inRecipe(&ing, w) // Checks if the ingredient is in a recipe
				if err != nil {
					http.Error(w, "Failed to check if ingredient is in recipe: "+err.Error(), http.StatusInternalServerError)
				}

				if !insideARecipe { // If it is not in a recipe, attempt to delete
					err = DBDelete(ing.ID, IngredientCollection, w)
					if err != nil {
						http.Error(w, "Failed to delete ingredient: "+err.Error(), http.StatusInternalServerError)
						return
					}

					fmt.Fprintln(w, "Successfully deleted ingredient "+ing.Name)
				} else {
					http.Error(w, "Can't delete ingredient"+ing.Name+" because it is used in a recipe.", http.StatusForbidden)
				}

			case caserec:
				rec := Recipe{}

				err := json.Unmarshal(resp, &rec)
				if err != nil {
					http.Error(w, "Could not unmarshal body of request"+err.Error(), http.StatusBadRequest)
					return
				}

				rec, err = DBReadRecipeByName(rec.RecipeName, w) //  Get that recipe
				if err != nil {
					http.Error(w, "Couldn't retrieve recipe: "+err.Error(), http.StatusBadRequest)
					return
				}

				err = DBDelete(rec.ID, RecipeCollection, w)
				if err != nil {
					http.Error(w, "Failed to delete recipe: "+err.Error(), http.StatusInternalServerError)
					return
				}

				fmt.Fprintln(w, "Successfully deleted recipe "+rec.RecipeName)
			}
		} else if err == nil {
			http.Error(w, "Not authorised to delete! Please use a valid token.", http.StatusUnauthorized)
		}
	default:
		http.Error(w, "Invalid method "+r.Method, http.StatusBadRequest)
	}
}

// RegisterIngredient func saves the ingredient to its respective collection in our firestore DB
func RegisterIngredient(w http.ResponseWriter, respo []byte) {
	ing := Ingredient{}
	found := false // ingredient found or not in database

	err := json.Unmarshal(respo, &ing)
	if err != nil {
		http.Error(w, "Could not unmarshal body of request"+err.Error(), http.StatusBadRequest)
		return
	}

	ing.Name = strings.ToLower(ing.Name) // force lowercase ingredient name

	if ing.Unit == "" {
		http.Error(w, "Could not save ingredient, missing \"unit\"", http.StatusBadRequest)
		return
	}

	unitParam := ing.Unit //  Checks if the posted unit is one of the legal measurements
	inList := false

	for _, v := range AllowedUnit { //  Loops through the allowed units
		if unitParam == v {
			inList = true
			break
		}
	} //  If it is one of the allowed units, cast it into g or l

	if inList {
		if strings.Contains(unitParam, "g") {
			unitParam = "g"
		} else if strings.Contains(unitParam, "l") {
			unitParam = "l"
		} else {
			unitParam = "pc"
		}
	} else { //  Prints the allowed units for an ingridient
		http.Error(w, "Unit has to be of one of the values ", http.StatusBadRequest)
		for _, v := range AllowedUnit {
			fmt.Fprintln(w, v) // Print allowed units
		}
		return
	}

	allIngredients, err := DBReadAllIngredients(w) // temporary list of all ingredients in database

	if err != nil {
		http.Error(w, "Could not retrieve collection "+IngredientCollection+" "+
			err.Error(), http.StatusInternalServerError)
		return
	}
	//  Check to see if the ingredient is already in the DB
	for i := range allIngredients {
		if ing.Name == allIngredients[i].Name {
			found = true // found ingredient in database

			http.Error(w, "Ingredient \""+ing.Name+"\" already in database.", http.StatusBadRequest)

			return
		}
	}

	if !found { // if ingredient is not found in database
		if unitParam != "pc" {
			ConvertUnit(&ing, unitParam) // convert unit to "g" or "l"
		}

		ing.Quantity = 1 // force quantity to 1

		err = GetNutrients(&ing, w) // get nutrients for the ingredient

		if err != nil {
			http.Error(w, "Couldn't get nutritional values: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if ing.Nutrients.Energy.Label == "" {
			// check if it got nutrients from db.
			//All ingredients will get this label if GetNutrients is ok
			http.Error(w, "ERROR: Failed to get nutrients for ingredient."+
				"Ingredient was not saved.", http.StatusInternalServerError)
		} else {
			err = DBSaveIngredient(&ing, w) // save it to database
			if err != nil {                 // if DBSaveIngredient return error
				http.Error(w, "Could not save document to collection "+
					IngredientCollection+" "+err.Error(), http.StatusInternalServerError)
			} else { // DBSaveIngredient did not return error
				err := CallURL(IngredientCollection, &ing, w) // Call webhooks
				if err != nil {
					fmt.Fprintln(w, "Could not post to webhooks.site: "+
						err.Error(), http.StatusBadRequest)
				}
				fmt.Fprintln(w, "Ingredient \""+ing.Name+"\" saved successfully to database.") // Success!
			}
		}
	}
}

// RegisterRecipe func saves the recipe to its respective collection in our firestore DB
func RegisterRecipe(w http.ResponseWriter, respo []byte) {
	rec := Recipe{}

	err := json.Unmarshal(respo, &rec)
	if err != nil {
		http.Error(w, "Could not unmarshal body of request"+err.Error(), http.StatusBadRequest)
		return
	}

	var missingingredients []string // name of ingredients in recipe missing in database

	//  Retrieve all recipes and ingredients to see if the one the user is trying to register already exists
	allRecipes, err := DBReadAllRecipes(w)
	if err != nil {
		http.Error(w, "Could not retrieve collection "+RecipeCollection+" "+
			err.Error(), http.StatusInternalServerError)
		return
	}
	//  Retrieves all the ingredients to get the ones missing for the recipe
	allIngredients, err := DBReadAllIngredients(w)
	if err != nil {
		http.Error(w, "Could not retrieve collection "+IngredientCollection+" "+
			err.Error(), http.StatusInternalServerError)
		return
	}
	//  If the name of the one created matches any of the ones in the DB
	for i := range allRecipes {
		if allRecipes[i].RecipeName == rec.RecipeName {
			http.Error(w, "Cannot save recipe, name already in use.", http.StatusBadRequest)
			return
		}
	}

	for i := range rec.Ingredients { // Loops through all the ingredients
		found := false // Reset if current ingredient is found or not

		for _, j := range allIngredients { // If the ingredient is found the loop breaks and found is set to true
			if rec.Ingredients[i].Name == j.Name {
				found = true

				// Check to see if user has posted with the equivalent unit as the ingredient has in the DB
				if !UnitCheck(rec.Ingredients[i].Unit, j.Unit) {
					//  Error message when posting with mismatched units, i.e liquid with kg or solid with ml
					http.Error(w, "Couldn't save recipe due to unit mismatch: "+
						rec.Ingredients[i].Name+" has unit "+j.Unit+
						" in database, and can not be saved with "+
						rec.Ingredients[i].Unit, http.StatusBadRequest)

					return
				}

				break
			}
		}

		if !found {
			missingingredients = append(missingingredients, rec.Ingredients[i].Name)
		}
	}

	//  If the ingredient found matches that of the recipe, the name is available and the unit of legal value
	if len(missingingredients) == 0 {
		err = GetRecipeNutrients(&rec, w) //  Collect the nutrients of that recipe

		if err != nil {
			http.Error(w, "Could not get nutrients for recipe", http.StatusInternalServerError)
			return
		}

		err = DBSaveRecipe(&rec, w) //  Saves the recipe

		if err != nil {
			http.Error(w, "Could not save document to collection "+
				RecipeCollection+" "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = CallURL(RecipeCollection, &rec, w) // Invokes the url

		if err != nil {
			http.Error(w, "Could not post to webhooks.site: "+err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Recipe \""+rec.RecipeName+"\" saved successfully to database.")
	} else {
		http.Error(w, "Cannot save recipe, missing ingredient(s) in database:", http.StatusBadRequest)

		for i := range missingingredients {
			fmt.Fprintln(w, "- "+missingingredients[i]) // print all missing ingredients in http response
		}
	}
}

// GetNutrients gets nutritional info from external API for the ingredient. Returns http error if it fails
func GetNutrients(ing *Ingredient, w http.ResponseWriter) error {
	client := http.DefaultClient

	APIURL := "http://api.edamam.com/api/nutrition-data?app_id="
	APIURL += AppID
	APIURL += "&app_key="
	APIURL += AppKey
	APIURL += "&ingr="
	// substitute spaces with "%20" so URL to API works with spaces in ingredient name
	APIURL += strings.ReplaceAll(ing.Name, " ", "%20")
	if ing.Unit != "pc" {
		APIURL += "%20"
		APIURL += ing.Unit
	} else if ing.Unit == "pc" {
		APIURL += "%20"
		APIURL += "piece"
	}

	resp, err := DoRequest(APIURL, client)

	if err != nil {
		http.Error(w, "Unable to get "+APIURL+err.Error(), http.StatusBadRequest)
		return err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	err = json.NewDecoder(resp.Body).Decode(&ing)

	if err != nil {
		http.Error(w, "Could not decode response body "+err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

// GetRecipeNutrients calculates total nutritients in a recipe
func GetRecipeNutrients(rec *Recipe, w http.ResponseWriter) error {
	// Set all the labels for the recipe
	rec.AllNutrients.Energy.Label = "Energy"
	rec.AllNutrients.Energy.Unit = "kcal"
	rec.AllNutrients.Fat.Label = "Fat"
	rec.AllNutrients.Fat.Unit = "g"
	rec.AllNutrients.Carbohydrate.Label = "Carbs"
	rec.AllNutrients.Carbohydrate.Unit = "g"
	rec.AllNutrients.Sugar.Label = "Sugar"
	rec.AllNutrients.Sugar.Unit = "g"
	rec.AllNutrients.Protein.Label = "Protein"
	rec.AllNutrients.Protein.Unit = "g"

	//  Loops through each ingredient in the recipe and adds up the nutritional information from each
	//  to a total amount of nutrients for the recipe as a whol
	for i := range rec.Ingredients {
		temptotalnutrients, err := CalcNutrition(rec.Ingredients[i], w)
		if err != nil {
			return err
		}

		rec.AllNutrients.Energy.Quantity += temptotalnutrients.Nutrients.Energy.Quantity
		rec.AllNutrients.Fat.Quantity += temptotalnutrients.Nutrients.Fat.Quantity
		rec.AllNutrients.Carbohydrate.Quantity += temptotalnutrients.Nutrients.Carbohydrate.Quantity
		rec.AllNutrients.Sugar.Quantity += temptotalnutrients.Nutrients.Sugar.Quantity
		rec.AllNutrients.Protein.Quantity += temptotalnutrients.Nutrients.Protein.Quantity

		rec.Ingredients[i].Nutrients.Energy = temptotalnutrients.Nutrients.Energy
		rec.Ingredients[i].Nutrients.Fat = temptotalnutrients.Nutrients.Fat
		rec.Ingredients[i].Nutrients.Carbohydrate = temptotalnutrients.Nutrients.Carbohydrate
		rec.Ingredients[i].Nutrients.Sugar = temptotalnutrients.Nutrients.Sugar
		rec.Ingredients[i].Nutrients.Protein = temptotalnutrients.Nutrients.Protein

		rec.Ingredients[i].Calories = temptotalnutrients.Nutrients.Energy.Quantity
		rec.Ingredients[i].ID = temptotalnutrients.ID
	}

	return nil
}

// inRecipe is a check to see if an ingredient is present in a recipe
func inRecipe(ing *Ingredient, w http.ResponseWriter) (bool, error) {
	//  Get all recipes
	recipes, err := DBReadAllRecipes(w) // Else get all recipes

	if err != nil {
		http.Error(w, "Couldn't retrieve recipes: "+err.Error(), http.StatusBadRequest)
		return false, err
	}

	for _, r := range recipes {
		for _, i := range r.Ingredients {
			if i.Name == ing.Name {
				return true, err
			}
		}
	}

	return false, err
}
