-- +migrate Up

SET ROLE vb_categories;

INSERT INTO categories (id, label, extensible, schema)
  VALUES ('venue-amenity', 'Venue amenities', TRUE,
'{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('venue-amenity', TRUE, 'car-park', '"car park"'),
         ('venue-amenity', TRUE, 'reception-24hr', '"24-hour reception"'),
         ('venue-amenity', TRUE, 'free-wifi', '"free WIFI"'),
         ('venue-amenity', TRUE, 'restaurant', '"restaurant"'),
         ('venue-amenity', TRUE, 'breakfast', '"breakfast"'),
         ('venue-amenity', TRUE, 'lunch', '"lunch"'),
         ('venue-amenity', TRUE, 'dinner', '"dinner"'),
         ('venue-amenity', TRUE, 'bar', '"bar"'),
         ('venue-amenity', TRUE, 'gym', '"gym"'),
         ('venue-amenity', TRUE, 'pool', '"swimming pool"'),
         ('venue-amenity', TRUE, 'sport-facilities', '"sports facilities"'),
         ('venue-amenity', TRUE, 'hot-tub', '"hot tub"'),
         ('venue-amenity', TRUE, 'group-friendly', '"good for groups"'),
         ('venue-amenity', TRUE, 'events-friendly', '"good for events"'),
         ('venue-amenity', TRUE, 'pets-allowed', '"pets allowed 🐾"'),
         ('venue-amenity', TRUE, 'smoking-allowed', '"smoking allowed"'),
         ('venue-amenity', TRUE, 'full-vegan', '"100% vegan <VB logo>"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('offer-category', 'Categories of offer', TRUE,
'{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('offer-category', TRUE, 'concert', '"Concert"'),
         ('offer-category', TRUE, 'class', '"Class"'),
         ('offer-category', TRUE, 'sports', '"Sports"'),
         ('offer-category', TRUE, 'health', '"Health"'),
         ('offer-category', TRUE, 'impact', '"Impact/environment"'),
         ('offer-category', TRUE, 'food', '"Food"'),
         ('offer-category', TRUE, 'travel', '"Travel"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('product-certification', 'Product certifications', FALSE,
'{ "type": "object", "properties": {"name": { "type": "string" }, "logo": { "type": "string", "format": "uri" }, "website": { "type": "string", "format": "uri" } }, "required": ["name"], "additionalProperties": false }');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES
('product-certification', TRUE, 'soil-association',
 '{"name": "Soil Association", "logo": "...", "website": "https://www.soilassociation.org/"}'),
('product-certification', TRUE, 'fair-trade',
 '{"name": "Fair Trade", "logo": "...", "website": "https://www.fairtrade.net/"}'),
('product-certification', TRUE, 'vegan-society',
'{"name": "Vegan Society", "logo": "...", "website": "https://www.vegansociety.com/"}');



INSERT INTO categories (id, label, extensible, schema)
  VALUES ('nutrition-info', 'Nutrition information', TRUE,
'{ "type": "object", "properties": {"name": { "type": "string" }, "parent": { "type": "string", "pattern": "^[a-z-]+$" }, "unit": { "type": "string" }, "has_rda": { "type": "boolean" } }, "required": ["name"], "additionalProperties": false }');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('nutrition-info', TRUE, 'total-fat',          '{"name": "Total fat"          , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'saturated-fat',      '{"name": "Saturated fat"      , "parent": "total-fat",         "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'trans-fat',          '{"name": "Trans fats"         , "parent": "total-fat",         "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'salt',               '{"name": "Salt"               , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'sodium',             '{"name": "Sodium"             , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'cholesterol',        '{"name": "Cholesterol"        , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'protein',            '{"name": "Protein"            , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'total-carbohydrate', '{"name": "Total carbohydrate" , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'dietary-fibre',      '{"name": "Dietary fibre"      , "parent": null,              "unit": "g", "has_rda": true}'),
         ('nutrition-info', TRUE, 'sugars',             '{"name": "Sugars"             , "parent": "total-carbohydrate","unit": "g", "has_rda": false}'),
         ('nutrition-info', TRUE, 'added-sugars',       '{"name": "Added sugars"       , "parent": "total-carbohydrate","unit": "g", "has_rda": false}'),
         ('nutrition-info', TRUE, 'calcium',            '{"name": "Calcium"            , "parent": null,               "unit": "?",  "has_rda": true}'),
         ('nutrition-info', TRUE, 'potassium',          '{"name": "Potassium"          , "parent": null,               "unit": "?",  "has_rda": true}'),
         ('nutrition-info', TRUE, 'iron',               '{"name": "Iron"               , "parent": null,               "unit": "?",  "has_rda": true}'),
         ('nutrition-info', TRUE, 'zinc',               '{"name": "Zinc"               , "parent": null,               "unit": "?",  "has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-a',          '{"name": "Vitamin A"          , "parent": null,              "unit": "μg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b1',         '{"name": "Vitamin B1"         , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b2',         '{"name": "Vitamin B2"         , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b3',         '{"name": "Vitamin B3"         , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b5',         '{"name": "Vitamin B5"         , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b6',         '{"name": "Vitamin B6"         , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b7',         '{"name": "Vitamin B7"         , "parent": null,              "unit": "μg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b9',         '{"name": "Vitamin B9"         , "parent": null,              "unit": "μg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-b12',        '{"name": "Vitamin B12"        , "parent": null,              "unit": "μg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-c',          '{"name": "Vitamin C"          , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-d',          '{"name": "Vitamin D"          , "parent": null,              "unit": "μg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-e',          '{"name": "Vitamin E"          , "parent": null,              "unit": "mg","has_rda": true}'),
         ('nutrition-info', TRUE, 'vitamin-k',          '{"name": "Vitamin K"          , "parent": null,              "unit": "μg","has_rda": true}');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('product-label', 'Product labels', TRUE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('product-label', TRUE, 'soy-free', '"Soy free"'),
         ('product-label', TRUE, 'gluten-free', '"Gluten free"'),
         ('product-label', TRUE, 'palm-oil-free', '"Palm-oil free"'),
         ('product-label', TRUE, 'nut-free', '"Nut free"'),
         ('product-label', TRUE, 'organic', '"Organic/bio"'),
         ('product-label', TRUE, 'wheat-free', '"Wheat free"'),
         ('product-label', TRUE, 'gmo-free', '"GMO free"'),
         ('product-label', TRUE, 'eco', '"Green/Eco"'),
         ('product-label', TRUE, 'toxin-free', '"Toxin free"'),
         ('product-label', TRUE, 'allergy-friendly', '"Allergy friendly"'),
         ('product-label', TRUE, 'cruelty-free', '"Cruelty free"'),
         ('product-label', TRUE, 'no-animal-testing', '"Not tested on animals"'),
         ('product-label', TRUE, 'gives-back', '"Gives back"'),
         ('product-label', TRUE, 'locally-sourced', '"Locally sourced"'),
         ('product-label', TRUE, 'shop-locally', '"Shop locally"'),
         ('product-label', TRUE, 'recycled', '"Recycled"'),
         ('product-label', TRUE, 'upcycled', '"Upcycled"'),
         ('product-label', TRUE, 'sustainable', '"Sustainable product/ingredients"'),
         ('product-label', TRUE, 'sustainable-packaging', '"Sustainable packaging"'),
         ('product-label', TRUE, 'peta-approved', '"PETA approved"'),
         ('product-label', TRUE, 'vegan-certified', '"Vegan certified"'),
         ('product-label', TRUE, 'supports-animal-sanctuaries', '"Supports animal sanctuaries"'),
         ('product-label', TRUE, 'women-owned-company', '"Women owned company"'),
         ('product-label', TRUE, 'vegan-owned-company', '"Vegan owned company"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('food-allergen', 'Food allergens', TRUE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('food-allergen', TRUE, 'balsam-of-peru', '"Balsam of Peru"'),
         ('food-allergen', TRUE, 'egg', '"egg"'),
         ('food-allergen', TRUE, 'fish', '"fish or shellfish"'),
         ('food-allergen', TRUE, 'fruit', '"fruit"'),
         ('food-allergen', TRUE, 'garlic', '"garlic"'),
         ('food-allergen', TRUE, 'hot-peppers', '"hot peppers"'),
         ('food-allergen', TRUE, 'oats', '"oats"'),
         ('food-allergen', TRUE, 'meat', '"meat"'),
         ('food-allergen', TRUE, 'milk', '"milk"'),
         ('food-allergen', TRUE, 'peanuts', '"peanuts"'),
         ('food-allergen', TRUE, 'rice', '"rice"'),
         ('food-allergen', TRUE, 'sesame', '"sesame"'),
         ('food-allergen', TRUE, 'soy', '"soy"'),
         ('food-allergen', TRUE, 'sulphites', '"sulphites"'),
         ('food-allergen', TRUE, 'tartrazine', '"tartrazine"'),
         ('food-allergen', TRUE, 'tree-nuts', '"tree nuts"'),
         ('food-allergen', TRUE, 'wheat', '"wheat"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('room-type', 'Room types', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('room-type', TRUE, 'single', '"single"'),
         ('room-type', TRUE, 'double', '"double"'),
         ('room-type', TRUE, 'twin', '"twin"'),
         ('room-type', TRUE, 'apartment', '"apartment"'),
         ('room-type', TRUE, 'apartment-with-kitchen', '"apartment with kitchen"'),
         ('room-type', TRUE, 'suite', '"suite"'),
         ('room-type', TRUE, 'chalet', '"chalet"'),
         ('room-type', TRUE, 'bungalow', '"bungalow"'),
         ('room-type', TRUE, 'cabin', '"cabin"'),
         ('room-type', TRUE, 'tent', '"tent"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('room-facility', 'Room facilities', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('room-facility', TRUE, 'private-bathroom', '"private bathroom"'),
         ('room-facility', TRUE, 'shower', '"shower"'),
         ('room-facility', TRUE, 'bath', '"bath"'),
         ('room-facility', TRUE, 'balcony', '"balcony"'),
         ('room-facility', TRUE, 'terrace', '"terrace"'),
         ('room-facility', TRUE, 'desk', '"desk"'),
         ('room-facility', TRUE, 'tv', '"TV"'),
         ('room-facility', TRUE, 'air-conditioning', '"air conditioning"'),
         ('room-facility', TRUE, 'telephone', '"telephone"'),
         ('room-facility', TRUE, 'free-wifi', '"free WiFi"'),
         ('room-facility', TRUE, 'hairdryer', '"hairdryer"'),
         ('room-facility', TRUE, 'wardrobe', '"wardrobe"'),
         ('room-facility', TRUE, 'towels', '"towels"'),
         ('room-facility', TRUE, 'linen', '"linen"'),
         ('room-facility', TRUE, 'clothes-rack', '"clothes rack"'),
         ('room-facility', TRUE, 'toilet-paper', '"toilet paper"'),
         ('room-facility', TRUE, 'cosmetics', '"cosmetics"'),
         ('room-facility', TRUE, 'tea-coffee-maker', '"tea/coffee maker"'),
         ('room-facility', TRUE, 'iron', '"iron"'),
         ('room-facility', TRUE, 'kitchenette', '"kitchenette"'),
         ('room-facility', TRUE, 'kitchen', '"kitchen"'),
         ('room-facility', TRUE, 'refrigerator', '"refrigerator"'),
         ('room-facility', TRUE, 'seating-area', '"seating area"'),
         ('room-facility', TRUE, 'dining-area', '"dining area"'),
         ('room-facility', TRUE, 'washing-machine', '"washing machine"'),
         ('room-facility', TRUE, 'heating', '"heating"'),
         ('room-facility', TRUE, 'cable-channels', '"cable channels"'),
         ('room-facility', TRUE, 'sofa', '"sofa"'),
         ('room-facility', TRUE, 'soundproofing', '"soundproofing"'),
         ('room-facility', TRUE, 'view', '"view"'),
         ('room-facility', TRUE, 'electric-kettle', '"electric kettle"'),
         ('room-facility', TRUE, 'kitchenware', '"kitchenware"'),
         ('room-facility', TRUE, 'hypoallergenic', '"hypoallergenic"'),
         ('room-facility', TRUE, 'cleaning-products', '"cleaning products"'),
         ('room-facility', TRUE, 'dining-table', '"dining table"'),
         ('room-facility', TRUE, 'childrens-high-chair', '"children’s high chair"'),
         ('room-facility', TRUE, 'childrens-crib', '"children’s crib"'),
         ('room-facility', TRUE, 'drying-rack', '"clothes drying rack"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('room-service', 'Room services', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('room-service', TRUE, 'daily-cleaning', '"daily cleaning"'),
         ('room-service', TRUE, 'towel-change', '"towel change"'),
         ('room-service', TRUE, 'linen-change', '"linen change"'),
         ('room-service', TRUE, 'luggage-storage', '"luggage storage"'),
         ('room-service', TRUE, 'laundry-service', '"laundry service"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('room-rule', 'Room rules', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('room-rule', TRUE, 'smoking-allowed', '"smoking allowed"'),
         ('room-rule', TRUE, 'pets-allowed', '"pets allowed"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('room-label', 'Room labels', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('room-label', TRUE, 'organic', '"organic"'),
         ('room-label', TRUE, 'minimalist', '"minimalist"'),
         ('room-label', TRUE, 'remote-work', '"ideal for remote work"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-entrance', 'Accessibility: entrance', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-entrance', TRUE, 'no-step-obstacle', '"no stairs or steps to enter"'),
         ('accessibility-entrance', TRUE, 'well-lit-path', '"well-lit path to entrance"'),
         ('accessibility-entrance', TRUE, 'wide-guest-entrance', '"wide entrance for guests"'),
         ('accessibility-entrance', TRUE, 'flat-path', '"flat path to guest entrance"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-movement', 'Accessibility: movement', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-movement', TRUE, 'wide-hallways', '"wide hallways"'),
         ('accessibility-movement', TRUE, 'elevator', '"elevator"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-bed', 'Accessibility: bed', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-bed', TRUE, 'accessible-height-bed', '"accessible-height bed"'),
         ('accessibility-bed', TRUE, 'extra-space-around-bed', '"extra space around bed"'),
         ('accessibility-bed', TRUE, 'electric-profiling-bed', '"electric profiling bed"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-bathroom', 'Accessibility: bathroom', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-bathroom', TRUE, 'no-step-obstacle', '"no stairs or steps to enter"'),
         ('accessibility-bathroom', TRUE, 'shower-grab-bars', '"fixed grab bars for shower"'),
         ('accessibility-bathroom', TRUE, 'toilet-grab-bars', '"fixed grab bars for toilet"'),
         ('accessibility-bathroom', TRUE, 'wide-doorway-to-bathroom', '"wide doorway to guest bathroom"'),
         ('accessibility-bathroom', TRUE, 'roll-in-shower', '"roll-in shower"'),
         ('accessibility-bathroom', TRUE, 'shower-chair', '"shower chair"'),
         ('accessibility-bathroom', TRUE, 'handheld-shower-head', '"handheld shower head"'),
         ('accessibility-bathroom', TRUE, 'bathtub-with-bath-chair', '"bathtub with bath chair"'),
         ('accessibility-bathroom', TRUE, 'accessible-height-toilet', '"accessible-height toilet"'),
         ('accessibility-bathroom', TRUE, 'shower-wide-clearance', '"wide clearance to shower"'),
         ('accessibility-bathroom', TRUE, 'toilet-wide-clearance', '"wide clearance to toilet"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-common-areas', 'Accessiblity: common areas', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-common-areas', TRUE, 'no-step-obstacle', '"no stairs or steps to enter"'),
         ('accessibility-common-areas', TRUE, 'wide-entryway', '"wide entryway"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-parking', 'Accessibility: parking', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-parking', TRUE, 'disabled-parking-space', '"disabled parking space"'),
         ('accessibility-parking', TRUE, 'wide-parking-space', '"parking space at least 2.5m (8 feet) wide"');


INSERT INTO categories (id, label, extensible, schema)
  VALUES ('accessibility-equipment', 'Accessibility: equipment', FALSE, '{"type": "string"}');

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('accessibility-equipment', TRUE, 'mobile-hoist', '"mobile hoist"'),
         ('accessibility-equipment', TRUE, 'pool-with-pool-hoist', '"pool with pool hoist"'),
         ('accessibility-equipment', TRUE, 'ceiling-hoist', '"ceiling hoist"');


-- +migrate Down

SET ROLE vb_categories;

DELETE FROM category_entries;
DELETE FROM categories;
