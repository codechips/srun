#!/bin/bash

# Function to show progress bar
show_progress() {
  local width=50
  local progress=$1
  local percent=$((progress * 100 / total_steps))
  local completed=$((width * progress / total_steps))
  local remaining=$((width - completed))

  printf "\r[%s%s] %d%%" "$(printf '#%.0s' $(seq 1 $completed))" "$(printf ' %.0s' $(seq 1 $remaining))" "$percent"
}

# Total number of steps
total_steps=100

echo "Starting long-running task..."

# Simulate work with a loop
for i in $(seq 1 $total_steps); do
  # Do some work (simulated with sleep)
  sleep 1

  # Update progress bar
  show_progress $i
done

echo -e "\nTask completed!"
