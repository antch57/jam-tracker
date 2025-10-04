# Jam Tracker

## Dev workflow

1. get db up:

```bash
docker-compose up -d
```

1. spin up app

```bash
go run cmd/api/main.go
```

## end goal..

This project will allow users to store information about shows they have been to. Users will be able to create profiles and log shows they have atteneded. The data about each show will store the band, venue, date, users rating, users favorite songs, or even show notes. Maybe the user can upload photos from the show... or ticket stubs. Maybe even have the option for friends so users can see what shows everyone is going to.

I would also like to get a recommendation system to suggest shows in the area that the user might want to attend. I think it would be awesome to leverage ML in this somehow.


copilot for the win.....

#### Show Data:

- Setlist.fm API (comprehensive setlist database)
- Songkick API (upcoming concerts)
- Bands In Town API (tour dates)
- Jambase.com scraping (jam band focus)

#### Geographic:

- Google Places API for venue data
- Distance calculations for location-based recommendations
- Development Phases
- MVP (Phase 1): Basic Show Tracking

#### User registration/auth

- Manual show entry with basic details
- Simple "users also liked" recommendations
- Web interface for show logging

#### Phase 2: Data Integration

- Automatic venue/band lookup via APIs
- Scraping upcoming shows in user's area
- Basic content-based filtering (genre matching)

#### Phase 3: ML Enhancement

- Implement collaborative filtering
- Add sophisticated feature engineering
- A/B testing for recommendation quality
- Mobile app or PWA

#### Phase 4: Advanced Features

- Social networking (friends' show recommendations)
- Playlist generation based on attended shows
- Integration with streaming services
- Predictive analytics (which tours to expect)
- Sample ML Features You Could Implement
- Band Similarity Graph: Build networks of band connections
- Venue Preference Learning: Learn if user prefers intimate vs. large venues
- Seasonal Pattern Recognition: Detect if user goes to more shows in summer
- Genre Evolution Tracking: How user's taste changes over time
- Friend Influence Modeling: Weight recommendations from friends' attendance

We could begin with:

Setting up the basic Go project structure
Designing the database schema
Building a simple web scraper for tour dates
Creating the basic REST API for show tracking
