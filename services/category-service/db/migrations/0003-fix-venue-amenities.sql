-- +migrate Up

SET ROLE vb_categories;

DELETE FROM category_entries WHERE category = 'venue-amenity';

INSERT INTO category_entries (category, fixed, label, value)
  VALUES ('venue-amenity', TRUE, 'car-park', '"Car park"'),
         ('venue-amenity', TRUE, 'reception-24hr', '"24-hour reception"'),
         ('venue-amenity', TRUE, 'free-wifi', '"Free WiFi"'),
         ('venue-amenity', TRUE, 'restaurant', '"Restaurant"'),
         ('venue-amenity', TRUE, 'breakfast', '"Breakfast"'),
         ('venue-amenity', TRUE, 'lunch', '"Lunch"'),
         ('venue-amenity', TRUE, 'dinner', '"Dinner"'),
         ('venue-amenity', TRUE, 'bar', '"Bar"'),
         ('venue-amenity', TRUE, 'gym', '"Gym"'),
         ('venue-amenity', TRUE, 'sauna', '"Sauna"'),
         ('venue-amenity', TRUE, 'spa', '"Spa & wellness"'),
         ('venue-amenity', TRUE, 'pool', '"Swimming pool"'),
         ('venue-amenity', TRUE, 'sport-facilities', '"Sports facilities"'),
         ('venue-amenity', TRUE, 'couples-friendly', '"Good for couples"'),
         ('venue-amenity', TRUE, 'family-friendly', '"Good for families"'),
         ('venue-amenity', TRUE, 'group-friendly', '"Good for groups"'),
         ('venue-amenity', TRUE, 'events-friendly', '"Good for events"'),
         ('venue-amenity', TRUE, 'pets-allowed', '"Pets allowed"'),
         ('venue-amenity', TRUE, 'full-vegan', '"100% vegan"');


-- +migrate Down

SET ROLE vb_categories;

DELETE FROM category_entries WHERE category = 'venue-amenity';

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
