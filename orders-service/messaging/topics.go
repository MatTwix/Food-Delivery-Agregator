package messaging

type Topic string

const (
	RestaurantCreatedTopic Topic = "restaurant.created"
	RestaurantUpdatedTopic Topic = "restaurant.updated"
	RestaurantDeletedTopic Topic = "restaurant.deleted"
)

var Topics = []Topic{
	RestaurantCreatedTopic,
	RestaurantUpdatedTopic,
	RestaurantDeletedTopic,
}
