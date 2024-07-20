table "mealplans" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = ALWAYS
    }
  }
  column "week" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.id]
  }
}
table "mealplans_recipes" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = ALWAYS
    }
  }
  column "mealplan_id" {
    null = true
    type = integer
  }
  column "recipe_id" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "mealplan_recipes_mealplan_id_fkey" {
    columns     = [column.mealplan_id]
    ref_columns = [table.mealplans.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "mealplan_recipes_recipe_id_fkey" {
    columns     = [column.recipe_id]
    ref_columns = [table.recipes.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "recipe_ingredients" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = ALWAYS
    }
  }
  column "recipe_id" {
    null = true
    type = integer
  }
  column "name" {
    null = true
    type = text
  }
  column "amount" {
    null = true
    type = text
  }
  column "calories" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "recipe_ingredients_recipe_id_fkey" {
    columns     = [column.recipe_id]
    ref_columns = [table.recipes.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "recipe_steps" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = ALWAYS
    }
  }
  column "text" {
    null = true
    type = text
  }
  column "order" {
    null = true
    type = integer
  }
  column "recipe_id" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "recipe_steps_recipe_id_fkey" {
    columns     = [column.recipe_id]
    ref_columns = [table.recipes.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
}
table "recipes" {
  schema = schema.public
  column "id" {
    null = false
    type = integer
    identity {
      generated = ALWAYS
    }
  }
  column "name" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "slug" {
    null = true
    type = text
  }
  primary_key {
    columns = [column.id]
  }
}
schema "public" {
  comment = "standard public schema"
}
