package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/influxdb/influxdb/client"
)

const (
	myHost        = "localhost"
	myPort        = 8086
	myDB          = "mydb"
	myMeasurement = "temperature"
)

func main() {

	u, err := url.Parse(fmt.Sprintf("http://%s:%d", myHost, myPort))
	if err != nil {
		log.Fatal(err)
	}

	conf := client.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	dur, ver, err := con.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Happy as a Hippo! %v, %s", dur, ver)

	writeRooms(con)
}

func writeRooms(con *client.Client) {
	var (
		roomname   = []string{"exec", "singleone", "singletwo", "doubleone", "doubletwo", "doublethree", "meetingone", "balcony"}
		sampleSize = (60 * 24 * 30)
		pts        = make([]client.Point, sampleSize)
	)

	rand.Seed(42)

	for j := 0; j < 8; j++ {
		for k := 0; k < 10; k++ {
			temp := rand.Intn(100)/10 + 20.0
			now := time.Now()
			secs := now.Unix() - (60 * 60 * 24 * 30)
			//			fmt.Printf("Creating %d %d\n", k, j)

			for i := 0; i < (60 * 24 * 30); i++ {
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

				pts[i] = client.Point{
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

				secs = secs + 60
			}
			//			fmt.Printf("Writing %d %d\n", k, j)
			bps := client.BatchPoints{
				Points:          pts,
				Database:        myDB,
				RetentionPolicy: "default",
			}
			_, err := con.Write(bps)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

}
