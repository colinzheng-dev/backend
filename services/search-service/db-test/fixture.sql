INSERT INTO item_full_text VALUES ('item0001', 'hotel', 'approved',
  setweight(to_tsvector('Chelsea Hotel'), 'A') ||
  setweight(to_tsvector('A nice hotel in Chelsea'), 'B') ||
  setweight(to_tsvector('cohen joplin new-york'), 'C') ||
  setweight(to_tsvector('The world-famous Chelsea Hotel, preferred haunt of celebrities...'), 'D'));

INSERT INTO item_full_text VALUES ('item0002', 'hotel', 'approved',
  setweight(to_tsvector('The Ritz'), 'A') ||
  setweight(to_tsvector('The premier hotel in London'), 'B') ||
  setweight(to_tsvector('london cream-tea'), 'C') ||
  setweight(to_tsvector('The best hotel in London (whatever they say at the Savoy!).'), 'D'));

INSERT INTO item_full_text VALUES ('item0003', 'hotel', 'approved',
  setweight(to_tsvector('Hotel California'), 'A') ||
  setweight(to_tsvector('You can check out any time you like, but you can never leave.'), 'B') ||
  setweight(to_tsvector('champagne colitas'), 'C') ||
  setweight(to_tsvector('On a dark desert highway, cool wind in my hair, Warm smell of colitas rising up through the air, My head grew heavy and my sight grew dim, I had to stop for the night.'), 'D'));

INSERT INTO item_full_text VALUES ('item0004', 'hotel', 'approved',
  setweight(to_tsvector('Dora''s Bed and Breakfast'), 'A') ||
  setweight(to_tsvector('Family-run, best cream teas in Devon.'), 'B') ||
  setweight(to_tsvector('cream-tea b-and-b'), 'C') ||
  setweight(to_tsvector('In a sleepy Devon village, we keep the tradition of the seaside B&B going.'), 'D'));


INSERT INTO item_locations VALUES ('item0001', 'hotel', 'approved', ST_MakePoint(-73.9994015, 40.7435059));
INSERT INTO item_locations VALUES ('item0002', 'hotel', 'approved', ST_MakePoint(-0.1437663, 51.5069446));
INSERT INTO item_locations VALUES ('item0003', 'hotel', 'approved', ST_MakePoint(13.3456575, 47.6527596));
INSERT INTO item_locations VALUES ('item0004', 'hotel', 'approved', ST_MakePoint(-4.27428, 51.1405433));
