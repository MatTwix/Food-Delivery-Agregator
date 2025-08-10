#!/bin/bash

# Reset Kafka consumer group offsets to avoid duplicate message processing

echo "Resetting Kafka consumer group offsets..."

# List of consumer groups to reset
CONSUMER_GROUPS=(
    "orders-service-restaurants-consumer-group"
    "orders-service-payments-consumer-group" 
    "orders-service-couriers-consumer-group"
)

# Topics to reset offsets for
TOPICS=(
    "restaurant.created"
    "restaurant.updated"
    "restaurant.deleted"
    "order.created"
    "order.paid"
    "order.delivered"
    "payment.succeeded"
    "payment.failed"
    "courier.assigned"
    "courier.search.failed"
)

# Reset offsets for each consumer group and topic
for group in "${CONSUMER_GROUPS[@]}"; do
    echo "Resetting offsets for consumer group: $group"
    for topic in "${TOPICS[@]}"; do
        echo "  Resetting topic: $topic"
        docker exec kafka kafka-consumer-groups \
            --bootstrap-server localhost:9092 \
            --group "$group" \
            --topic "$topic" \
            --reset-offsets \
            --to-latest \
            --execute 2>/dev/null || echo "    Topic $topic not found or already reset"
    done
    echo ""
done

echo "Kafka consumer group offsets reset complete!"
echo ""
echo "You can now restart your services to avoid duplicate message processing."
