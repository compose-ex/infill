package main

import (
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client"
)

func main() {
	myURL := os.Getenv("INFLUX_URL")
	if myURL == "" {
		myURL = "localhost"
	}

	myDB := os.Getenv("INFLUX_DB")
	if myDB == "" {
		myDB = "mydb"
	}

	u, err := url.Parse(myURL)
	if err != nil {
		log.Fatal(err)
	}

	conf := client.Config{
		URL: *u,
		// Username: os.Getenv("INFLUX_USER"),
		// Password: os.Getenv("INFLUX_PWD"),
	}

	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	dur, ver, err := con.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected in %v to a InfluxDB version %s", dur, ver)

	writeRooms(con, myDB)
}

func writeRooms(con *client.Client, myDB string) {
	var (
		roomname   = []string{"exec", "singleone", "singletwo", "doubleone", "doubletwo", "doublethree", "meetingone", "balcony"}
		sampleSize = 10000
		pts        = make([]client.Point, sampleSize)
	)

	rand.Seed(42)

	// We'll track where we are building a batch of points with the batchptr
	batchptr := 0

	// Let's start by looping through all the rooms then through each floor.
	for j := 0; j < 8; j++ {
		for k := 0; k < 10; k++ {

			// Now we'll pick a "starting temperature"
			temp := rand.Intn(100)/10 + 20.0

			// This all starts from a month ago, so we'll find now and take a month off
			now := time.Now()
			secs := now.Unix() - (60 * 60 * 24 * 30)

			// Now we travel forward in time back to now...
			for i := 0; i < (60 * 24 * 30); i++ {
				// We decide on our temperature behavior here. This uses a switch with
				// a random number in case we want to make the behavior richer/more
				// erratic. For now, most of the time, the temperature goes up or down by
				// up to .5 degrees

				switch rand.Intn(10) {
				case 0, 1, 2, 3, 4, 5, 6:
					tchange := rand.Intn(11) - 5
					temp = temp - (tchange / 10)
					if temp < 0 {
						temp = 0.0
					}
					if temp > 40 {
						temp = 40.0
					}
				default:
				}

				// New we can create the point of data at batchptr. These all go in a
				// We fill in the room with a name from our array above, set the
				// level and set our created timestamp.

				pts[batchptr] = client.Point{
					Measurement: "temperature",
					Tags: map[string]string{
						"room":  roomname[j],
						"level": strconv.Itoa(k),
					},
					Fields: map[string]interface{}{
						"temp": temp,
					},
					Time:      time.Unix(secs, 0),
					Precision: "s",
				}

				// We increment the pointer now...
				batchptr = batchptr + 1

				// And if we hit the sampleSize its time to send
				if batchptr == sampleSize {
					// We create a BatchPoints structure with our points, set the db
					// it is heading for...
					bps := client.BatchPoints{
						Points:          pts,
						Database:        myDB,
						RetentionPolicy: "default",
					}

					// And then we write it
					_, err := con.Write(bps)

					// If that fails, we bomb out here with an error
					if err != nil {
						log.Fatal(err)
					}

					// Otherwise we reallocate the pts (letting the GC clean up) and
					// reset the batchptr
					pts = make([]client.Point, sampleSize)
					batchptr = 0
				}

				// Now we bump the clock on 60 seconds
				secs = secs + 60
			}
		}
	}

	// Tidy up by copying the incomplete batch to the server
	if batchptr > 0 {
		newpts := make([]client.Point, batchptr)
		copy(newpts, pts)

		bps := client.BatchPoints{
			Points:          newpts,
			Database:        myDB,
			RetentionPolicy: "default",
		}
		_, err := con.Write(bps)
		if err != nil {
			log.Fatal(err)
		}
	}
}
