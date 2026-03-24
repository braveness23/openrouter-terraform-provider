# Retrieve all activity from the last 30 days
data "openrouter_activity" "all" {}

# Retrieve activity for a specific day
data "openrouter_activity" "today" {
  date = "2026-03-24"
}

output "total_requests_today" {
  value = sum([for row in data.openrouter_activity.today.activity : row.requests])
}
