# CloudProject
NTNU Cloud Technologies project 2019

The full project task description can be found further below. 
In short, the task was to make a service where you can get food recipes via a REST API, using ingredients you already have at your disposal. The nutritional facts for the recipes will also be automatically fetched. It is possible to add more or delete ingredients and recipes from the database (requires token). 

# Handlers: 
## Food 
Description: 
This handler lets you register and delete either an ingrededient or a recipe to the database. You have to post the body in a json format, and an access token is required.

### Register ingredient: cravings/food/ingredient
	
	Register Ingredient: 
	{
		"token":"",
		"name":"",
		"unit":""
	}

Unit should be either "l" or "g".

	Example ingredient: 
	{
		"token":"YourToken",
		"name":"cinnamon",
		"unit":"g"
	}

### Register recipe: cravings/food/recipe
	
	Register Recipe:
	{
		"token":"",
		"recipeName":"",
		"ingredients":[
			{
				"name":"",
				"quantity":,
				"unit":""
			},
			{
				"name":"",
				"quantity":,
				"unit":""
			}
		],
		"description":[
				"",
				"",
				"",
				""
		]
	}

Each string will have its own line automatically. No linebreaks are needed in the strings.

	Example recipe: 
	{
		"token":"YourToken",
		"recipeName":"The Best Recipe Ever Made",
		"ingredients":[
			{
				"name":"milk",
				"quantity":1,
				"unit":"dl"
			},
			{
				"name":"flour",
				"quantity":500,
				"unit":"g"
			},
			{
				"name":"butter",
				"quantity":200,
				"unit":"g"
			},
			{
				"name":"cinnamon",
				"quantity":2,
				"unit":"teaspoon"
			}
		],
		"description":[
				"Mix it good.",
				"Make it tight.",
				"Bake the food.",
				"Eat at night."
		]
	}

### Delete ingredient or recipe
	Send a DELETE request to either: 
	cravings/food/ingredient
	cravings/food/recipe


	Delete Ingredient: 
	{
		"token":"",
		"name":""
	}

	Delete Recipe: 
	{
		"token":"",
		"recipeName":""
	}

The difference where ingredient has "name" and recipe has "recipeName" is to prevent confusion and accidents.

## HandlerMeal

Description: 
Handler meal is a handler which the user can write in whatever ingredients they have (further instructions below),
and the program will return a list of recipes. Depending on the 'allowMissing' parameter, the recipes will suggest what the user
could potentially make, or if it is set to false, the recipes will only appear if you have all the ingredients. 
Furthermore, the user will be able to see what ingredients are remaining after posting. 
	mealHandler:
		Get method:
			URL: /cravings/meal
			example for one ingredient: /cravings/meal/?ingredients=milk|2|l
			example for multiple	  : /cravings/meal/?ingredients=milk|2|l_tomato|4|kg_cardamom|500|g

			'_' splits up the different ingredients
			'|' splits up the ingredient, quantity and unit (in this given order)

			example with sortBy and allowMissing: /cravings/meal/?ingredients=milk|2|l&sortBy=have&allowMissing=false
			Default value = *
			sortBy(Optional): missing*, have, remaining	
			allowMissing(Optional): false*, true

		Post method:
	[	
		{
			"name": "ingredient name",
			"unit": "ingredient unit",
			"quantity": ingredient quantity
		}
	]
	
	Example:
	[
		{
			"name": "milk",
			"unit": "l",
			"quantity": 2	
		},
		{
			"name": "tomato",
			"unit": "kg",
			"quantity": 4
		},
		{
			"name": "cardamom",
			"unit": "g",
			"quantity": 500
		}
	]
list as many ingredients with quantity and unit as you want

The user can send a post request with the payload of the 'remaining' struct of any given recipe to get the recipe for 'the next meal'. This process can be done repeatedly until the 'remaining' list is empty.

	limit: int, sets to 5 as default
	allowMissing: bool, true as default. Decides wether or not to print out recipes that are missing ingredients
	sortBy: "have"|"missing"|"remaining". have sorts in a descending order, missing and remaining sorts in an ascending order

# Webhooks
Webhooks endpoint: /cravings/webhooks/
Here you can get information about webhooks for this website

Post method:
Post is used to create new webhooks.
Use endpoint:
/cravings/webhooks/

	And send with body:
	{
	"event":"[Event name]",
	"url":"[Url name]"
	}

Get method:
Get is used to see all or one choosen webhook.
To get all webhooks, use normal endpoint:
/cravings/webhooks/

To get one webhook, use normal endpoint + choosen id for webhook:
/cravings/webhooks/[ID]

Delete method:
Delete is used to delete one webhook.
Use endpoint:
/cravings/webhooks/

	And send with body:
	{
	"id":"[ID]"
	}

# Test
Test cover = 76,0%
Test coverage can be tested by entering following command in terminal: go test -cover

# Docker
A Dockerfile is included in this repository. This is tested to work with our build and the following commands. 

Example command for building docker image: 

	docker build -t cravings .

(working directory should be in the repository when executing the build command)

Example command for running the container: 

	docker run -i -t -p 8081:8080 cravings

(this will run it on port 8081 on the host machine)


# Original project plan
We will make an API that can be used to get meal ideas from what ingredients you already have. This API could for example be used by a website or app providing a GUI to the users.

We have a database containing ingredients and recipes, including nutritional info. Each request to our API reads data from the database. When registering new recipes or ingredients to the database, it will get the nutritional info from an external API (Edamam). 
The project will use both OpenStack and Docker, and store data in Firebase. 

Example usage: 
You post what ingredients you already have, including how much of each ingredient. In return you get recipes you can make using these ingredients, and also nutritional info about it (calories, fat, carbohydrates etc.). 

Another example:
User can read every ingredient/recipe in the database including its nutrients. 
Furthermore, user can get one recipe or ingredient by name

Getting recipes:
What ingredients you have is provided either via URL or preferrably using POST request with a JSON body. Recipes that can be made using the ingredients you already have will be returned. These will also include nutritional info for the recipe. The user also get what he/she has left after using a recipe, and which recipes you also can make afterwards without having to buy new ingredients.

Registration: 
You will need an auth token provided by us to get access to register and delete recipes or ingredients to the database. Tokens are stored in our database in a separate collection. 

Registration is done by sending a POST request to our registration handler for ingredients or recipe, including a JSON structure in body. We will provide templates for this. If this is used from an app or website with GUI, this JSON structure will not be shown to the end users, but rather the developers of the app/website to make functionality in the GUI, and probably autofill it from some text fields etc.

When a new ingredient is registered, we get the nutritional info for it from the Edamam API. New recipes get its nutritional info calculated from our database to avoid hitting any limits on the Edamam API. The only exception is recipes having ingredients specified with the unit "teaspoon" or "tablespoon". Then the ingredient with spoons as unit is checked against Edamam, but all the rest is still calculated from our database. 

Webhooks: 
Webhooks for seeing what’s registered into the database through the /register/ handler. This includes both recipes and ingredients.

Statistics: 
Statistics indicates the availability of the database used in the assignment, and the website of which the program retrieves information.  In addition, it indicates time elapsed since the start of the program. Last but not least, it indicates how many recipes and ingredients are stored in the database. 

Potential expansions of this project:
#1 Possibility to register several recipes in one single POST-request.
We chose not to implement this as we don't feel it is important. It shouldn't be too hard though. We could make a small function which splits the recipes in the JSON body and send them individually with the code we already have. 

#2 Registration of recipes and ingredients could be done automatically via some external API or website. 
We chose not to implement this since many recipe databases are copyrighted, and we want quality over quantity. We didn't need to spend much time to register some example recipes ourselves, which is enough to show how our program works. If this program were to be used in a real setting with users, we would want to fill our database with good recipes, and allow submission of recipes from users, maybe with some application form and moderation, thus needing an access token.

#3 User requests a recipe, inserts what it has of ingredients. The system provides a “shopping list”.
A basic shopping list would be unnessecary since you could just print the recipe itself - that is basically the shopping list. We thought about making an option to fill your shopping cart at a web store automatically with the ingredients your need at for instance www.kolonial.no. This could actually be useful for Norwegian citizens. They have not made their API open yet since it is still under development.


# What went well and wrong 
After about an hour into our first meeting we had layed out a project plan of what our final product should look like. We had good working routines, meeting as a group every day to work together.

We managed to reach all of our main goals, and added some of the potential expansions for the project.

When the user 'uses' a recipe, there is a list of all ingredients he/she has for the recipe(have), what he/she needs to complete the recipe(missing) and ingredients after making the recipe(remaining). To find recipes for the next day, the application just posts a new request with the remaining ingredients list.
We also do different queries to handlerMeal (look at #handlerMeal for more information)


## Hard aspects of the project
Recipes that has units in teaspoon or tablespoon values became a bigger problem fixing than expected, since ingredients are saved in grams or litres in the database. This would not have been a problem if we simply could calculate all ingredients by volume, but we don't know how many grams x volume of each ingredient is. We did not want to enforce recipes to use weight instead of spoons, so to go around this, the nutritional value for each spoon when registering a recipe is being checked against the external API.
The solution we decided to go for in the meal-handler was calculating how many calories there was per spoon and from there get the quantity per unit. This lead to extra lines of code only for handling spoon units, but we still managed to only read from our own database every time a recipe is read, and we avoid storing duplicate ingredients with different units.


## What we learned
We got a deeper insight in how it is to make an API with a database, that is meant to be used by other applications. After a considerable amount of work was done we found several bugs and inconsistencies that were problematic for the program to run as we pictured. As a consequence we spent many hours debugging and figuring out how to handle errors in a greater scale compared to the knowledge we had in the initial phases of this project. 

From this project we were able to get a deeper understanding of how to use json and how the different methods of encoding and decoding the payload works. For instance, when we wanted to use json.NewDecoder to read a json body into a struct, we found out that it did not work because it reads from an empty file the second time. Therefore, we had to use unmarshal for our functions to work. 
In conclusion we got a better understanding of everything we have learnt so far in this course. 
