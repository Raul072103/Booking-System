go build -o bookings.exe ./cmd/web/.
bookings.exe -dbname=postgres -dbuser=postgres -cache=false -production=false