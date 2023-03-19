package iot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iot-mqtt-api/db"
	"iot-mqtt-api/dbstructs"
	"iot-mqtt-api/jsondecodebody"
	"iot-mqtt-api/logger"
	"iot-mqtt-api/mqtt"
	"iot-mqtt-api/openweatherdata"
	"iot-mqtt-api/user"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type IOT struct {
	database *db.Database
	mqtt     *mqtt.MQTT
	log      *logger.Logger
	data     Data
	// user     user.User
	owDataURL string
	mr        jsondecodebody.MalformedRequest
}

type Data struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Pressure    float64 `json:"pressure"`
	RelayOn     bool    `json:"RelayOn"`
}

func NewIOT(d *db.Database, m *mqtt.MQTT, l *logger.Logger, owDataURL string) *IOT {
	return &IOT{database: d, mqtt: m, log: l, owDataURL: owDataURL}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func (iot *IOT) WebServer() {
	// available URL-s for actions to be executed
	// get data from ESP8266
	http.HandleFunc("/bme280/getData", iot.getData)

	// get Open weather data
	http.HandleFunc("/getOpenWeatherData", iot.getOpenWeatherData)

	// manipulate relay
	http.HandleFunc("/setRelayStatus", iot.setReleyStatus)

	// create new data
	http.HandleFunc("/newUser", iot.newUser)
	http.HandleFunc("/newBuilding", iot.newBuilding)
	http.HandleFunc("/newFloor", iot.newFloor)
	http.HandleFunc("/newRoom", iot.newRoom)
	http.HandleFunc("/newDevice", iot.newDevice)
	http.HandleFunc("/addDeviceToUser", iot.addDeviceToUser)
	http.HandleFunc("/setDesiredTemperatureToDevice", iot.setDesiredTemperatureToDevice)

	// authenticate user
	http.HandleFunc("/authUser", iot.authenticateUser)

	// get data
	http.HandleFunc("/getUserBuildingsAndDevices", iot.getUserBuildingsAndDevices)
	http.HandleFunc("/getDeviceData", iot.getDeviceData)
	http.HandleFunc("/getHeatingHistory", iot.getHeatingHistory)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (iot *IOT) handleError(w http.ResponseWriter, r *http.Request, errId int, errMsg string) {
	iot.mr.Status = errId
	iot.mr.Msg = errMsg
	http.Error(w, iot.mr.Msg, http.StatusInternalServerError)
	iot.log.Error(fmt.Sprint("\nError ID: ", iot.mr.Status, "\nError Message: ", iot.mr.Msg))
	iot.mr.Msg = ""
	iot.mr.Status = 0
}

func (iot *IOT) ReceiveData() {
	for {
		message := <-iot.mqtt.Messages
		messageString := message.Topic()
		deviceId, _ := strconv.Atoi(strings.Split(messageString, "/")[7])
		data := strings.Split(messageString, "/")[8]

		switch data {
		case "temp":
			iot.data.Temperature, _ = strconv.ParseFloat(string(message.Payload()), 64)
			iot.insertTemperatureInDB(deviceId)
			fmt.Printf("%s - %s°C\n", data, message.Payload())
		case "humidity":
			iot.data.Humidity, _ = strconv.ParseFloat(string(message.Payload()), 64)
			fmt.Printf("%s - %s%s\n", data, message.Payload(), "%")
			iot.insertHumidityInDB(deviceId)
		case "pressure":
			iot.data.Pressure, _ = strconv.ParseFloat(string(message.Payload()), 64)
			// fmt.Printf("%s - %shPa\n", data, message.Payload())
		case "relayOn":
			iot.data.RelayOn, _ = strconv.ParseBool(string(message.Payload()))
			// fmt.Printf("%s - %s\n", data, message.Payload())
		}
	}
}

// returns data as an json object
func (iot *IOT) getData(w http.ResponseWriter, r *http.Request) {
	jsonData, _ := json.Marshal(iot.data)
	fmt.Fprintf(w, "%s", jsonData)
}

func (iot *IOT) authenticateUser(w http.ResponseWriter, r *http.Request) {

	var requestData user.User

	// decoder := json.NewDecoder(r.Body)
	// err := decoder.Decode(&requestData)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(requestData)

	decErr := jsondecodebody.DecodeJSONBody(w, r, &requestData)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	fmt.Println("user", &requestData)

	var usr user.User
	rows, err :=
		iot.database.Conn.Query(`
		SELECT
			ID, FIRST_NAME, LAST_NAME, GENDER, USERNAME, PASSWORD, EMAIL, BIRTH_DATE, CREATED, UPDATED, CREATED_BY, UPDATED_BY
		FROM
			USERS 
		WHERE 
			(USERNAME = $1 OR EMAIL = $2) AND PASSWORD = $3`,
			requestData.Username, requestData.Username, requestData.Password)
	if err != nil {
		iot.log.Error(err.Error())
	}

	for rows.Next() {
		rows.Scan(&usr.Id, &usr.FirstName, &usr.LastName, &usr.Gender, &usr.Username, &usr.Password, &usr.Email, &usr.BirthDate, &usr.Created, &usr.Updated, &usr.CreatedBy, &usr.UpdatedBy)
		fmt.Printf("got: Id: %v, firstname: %v, lastname %v, gender: %v, username: %v, pwd: %v, email: %v, bday: %v\n",
			usr.Id, usr.FirstName, usr.LastName, usr.Gender, usr.Username, usr.Password, usr.Email, usr.BirthDate)
	}

	if usr.Id == 0 {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	} else {
		jsonData, _ := json.Marshal(usr)
		fmt.Fprintf(w, "%s", jsonData)
	}
}

func (iot *IOT) setReleyStatus(w http.ResponseWriter, r *http.Request) {
	var mqttPath mqtt.MQTTPathToDevice
	decErr := jsondecodebody.DecodeJSONBody(w, r, &mqttPath)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	var path = fmt.Sprintf("iot/%v/%v/floor%v/room/%v/dev/%v/relayOn", mqttPath.UserName, mqttPath.BuildingName, mqttPath.FloorId, mqttPath.RoomId, mqttPath.DeviceId)
	enableCors(&w)
	iot.mqtt.Publish(path, mqttPath.RelayOn)
	jsonData, _ := json.Marshal(mqttPath)
	fmt.Fprintf(w, "%s", jsonData)
	fmt.Println("Relay for heating", mqttPath.RelayOn)
}

func (iot *IOT) getOpenWeatherData(w http.ResponseWriter, r *http.Request) {

	var cityMap openweatherdata.City
	decoder := jsondecodebody.DecodeJSONBody(w, r, &cityMap)
	if decoder != nil {
		fmt.Println(decoder.Error())
	}
	r.Body.Close()
	url := iot.owDataURL + "&q=" + cityMap.City

	resp, getErr := http.Get(url)
	if getErr != nil {
		fmt.Println(getErr.Error())
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	data := openweatherdata.OpenWeatherData{}
	unrmashallErr := json.Unmarshal([]byte(responseData), &data)

	if unrmashallErr != nil {
		log.Fatal(unrmashallErr)
	}

	var dataForApp openweatherdata.DataForApp
	dataForApp.WeatherStatus = data.Weather[0].Main
	dataForApp.Icon = data.Weather[0].Icon
	dataForApp.Temperature = data.Main.Temperature
	jsonData, _ := json.Marshal(dataForApp)
	fmt.Fprintf(w, "%s", jsonData)
}

func (iot *IOT) newUser(w http.ResponseWriter, r *http.Request) {
	var user user.User
	decErr := jsondecodebody.DecodeJSONBody(w, r, &user)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		_, err :=
			iot.database.Conn.Exec(`
			INSERT INTO USERS 
				(FIRST_NAME, LAST_NAME, GENDER, USERNAME, PASSWORD, EMAIL, BIRTH_DATE, CREATED, UPDATED, CREATED_BY, UPDATED_BY)
			VALUES
				($1, $2, $3, $4, $5, $6, $7, GETDATE(), GETDATE(), $10, $11)`,
				user.FirstName, user.LastName, user.Gender, user.Username, user.Password, user.Email, user.BirthDate, nil, nil, user.CreatedBy, user.UpdatedBy)
		if err != nil {
			iot.log.Error(err.Error())
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "Person: %+v", user)
		}
	}
}

func (iot *IOT) newBuilding(w http.ResponseWriter, r *http.Request) {
	var building dbstructs.Buildings
	decErr := jsondecodebody.DecodeJSONBody(w, r, &building)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		rows, sqlErr :=
			iot.database.Conn.Query(`
			BEGIN
				declare @adress as nvarchar(50)
			SELECT
				@adress =
					(
					SELECT TOP 1
						CONCAT(STREET, ' ', STREET_NUMBER)
					FROM
						BUILDINGS
					WHERE
						STREET = $1 AND
						STREET_NUMBER = $2
					)
				IF @adress IS NULL
					BEGIN
						INSERT INTO BUILDINGS (NAME, STREET, STREET_NUMBER, COUNTRY, CITY, ZIP_CODE, LOCATION, CREATED_BY, UPDATED_BY)
						VALUES ($3, $4, $5, $6, $7, $8, $9, 1, 1)
				   RETURN
				END
				ELSE
					BEGIN
						SELECT ID = -1
					RETURN
				END
			END`,
				building.Street, building.StreetNumber,
				building.Name, building.Street, building.StreetNumber, building.Country, building.City, building.ZipCode, building.Location)
		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
			fmt.Printf("got: Id: %v", mr.Status)
		}

		if sqlErr != nil {
			iot.handleError(w, r, -500, sqlErr.Error())
		} else if mr.Status == -1 {
			iot.handleError(w, r, mr.Status, fmt.Sprint("Building with adress ", building.Street, " ", building.StreetNumber, " already exists "))
		} else {
			fmt.Fprintf(w, "Building: %+v", building)
		}
	}
}

func (iot *IOT) newFloor(w http.ResponseWriter, r *http.Request) {
	var floor dbstructs.Floors
	decErr := jsondecodebody.DecodeJSONBody(w, r, &floor)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		rows, sqlErr :=
			iot.database.Conn.Query(`
			BEGIN
				declare @floor as int
				declare @buildingId as int
			SELECT 
				@buildingId = (
					SELECT
						ID
					FROM
						BUILDINGS
					WHERE
						ID = $1
				)
			IF @buildingId IS NULL
				BEGIN
					SELECT
					ERROR_ID = -2
				RETURN
				END
			ELSE
				SELECT
					@floor =
						(
						SELECT 
							ID
						FROM
							FLOORS
						WHERE
							BUILDING_ID = $2 AND
							FLOOR = $3
						)
					IF @floor IS NULL
						BEGIN
							INSERT INTO FLOORS (BUILDING_ID, FLOOR, DEVICES, CREATED_BY, UPDATED_BY)
							VALUES ($4, $5, $6, 1, 1)
					RETURN
					END
					ELSE
						BEGIN
							SELECT ERROR_ID = -3
						RETURN
					END
				END`,
				floor.BuildingId,
				floor.BuildingId, floor.Floor,
				floor.BuildingId, floor.Floor, floor.Devices)

		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
		}
		if mr.Status == -2 {
			iot.handleError(w, r, mr.Status, fmt.Sprint("Building with id ", floor.BuildingId, " does not exist"))
		} else if mr.Status == -3 {
			iot.handleError(w, r, mr.Status, floor.FloorExistsError())
		} else {
			floor.ID = uint(mr.Status)
			if sqlErr != nil {
				iot.handleError(w, r, -4, sqlErr.Error())
			} else {
				fmt.Fprintf(w, "Floor: %+v", floor)
			}
		}
	}
}

func (iot *IOT) newRoom(w http.ResponseWriter, r *http.Request) {
	var room dbstructs.Rooms
	decErr := jsondecodebody.DecodeJSONBody(w, r, &room)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		rows, sqlErr :=
			iot.database.Conn.Query(`
		BEGIN
			declare @roomName as int
			declare @floorId as int
		SELECT
			@floorId = (
			SELECT
				f.ID
			FROM
				FLOORS as f
			LEFT JOIN 
				BUILDINGS as b ON f.BUILDING_ID = b.ID
			WHERE
				f.ID = $2
			)
		IF @floorId IS NULL
			BEGIN
				SELECT
					ERROR_ID = -4
				RETURN
				END
		ELSE
			SELECT
				@roomName =(
					SELECT
						ID
					FROM
						ROOMS
					WHERE
						FLOOR_ID = $3 AND
						ROOM_NAME = $4
				)
			IF @roomName IS NULL
				BEGIN
					INSERT INTO ROOMS (FLOOR_ID, ROOM_NAME, CREATED_BY, UPDATED_BY)
					VALUES (@floorId, $5, 1, 1)
				RETURN
				END
			ELSE
				BEGIN
					SELECT ERROR_ID = -5
				RETURN
			END
		END`,
				room.BuildingId, room.FloorId,
				room.FloorId, room.RoomName,
				room.RoomName)

		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
		}
		if mr.Status == -4 {
			iot.handleError(w, r, mr.Status, "no floorId")
		} else if mr.Status == -5 {
			iot.handleError(w, r, mr.Status, "room name exists")
		} else {
			room.ID = uint(mr.Status)
			if sqlErr != nil {
				iot.log.Error(sqlErr.Error())
				http.Error(w, fmt.Sprint(sqlErr), http.StatusInternalServerError)
			} else {
				fmt.Fprintf(w, "Room: %+v", room)
			}
		}
	}
}

func (iot *IOT) newDevice(w http.ResponseWriter, r *http.Request) {
	var device dbstructs.Devices
	decErr := jsondecodebody.DecodeJSONBody(w, r, &device)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		rows, sqlErr :=
			iot.database.Conn.Query(`
			BEGIN
			declare @roomId as int
			declare @roomDevices as int
		SELECT
			@roomId = (
			SELECT
				r.ID
			FROM
				ROOMS as r
			WHERE
				r.ID = $1
			),
		@roomDevices = (
			SELECT
				r.DEVICES
			FROM
				ROOMS as r
			WHERE
				r.ID = $1
			)
		IF @roomId IS NULL
			BEGIN
				SELECT
					ERROR_ID = -6
				RETURN
				END
		ELSE
			BEGIN
				INSERT INTO DEVICES(ROOM_ID, IS_CONTROL_DEVICE, IS_CONTROL_ON, IS_DEVICE_ACTIVE, CREATED_BY, UPDATED_BY)
				VALUES (@roomId, $2, $3, $4, 1, 1)
				UPDATE
					ROOMS
				SET
					DEVICES = @roomDevices + 1
			RETURN
			END
		END`,
				device.RoomId,
				device.IsControlDevice, device.IsControlOn, device.IsDeviceActive)

		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
		}
		if mr.Status == -6 {
			iot.handleError(w, r, mr.Status, "room with given ID does not exist")
		} else {
			device.ID = uint(mr.Status)
			if sqlErr != nil {
				iot.log.Error(sqlErr.Error())
				http.Error(w, fmt.Sprint(sqlErr), http.StatusInternalServerError)
			} else {
				fmt.Fprintf(w, "device: %+v", device)
			}
		}
	}
}

func (iot *IOT) addDeviceToUser(w http.ResponseWriter, r *http.Request) {
	var userDevices *dbstructs.UserDevices
	decErr := jsondecodebody.DecodeJSONBody(w, r, &userDevices)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		rows, sqlErr :=
			iot.database.Conn.Query(`
			BEGIN
			declare @deviceId as int
			declare @userId as int
			declare @userDeviceExists as bit
		SELECT
			@userDeviceExists = (
				SELECT
					ID
				FROM
					USER_DEVICES
				WHERE
					USER_ID = $1 AND
					DEVICE_ID = $2
			),
			@deviceId = (
				SELECT
					ID
				FROM
					DEVICES
				WHERE
					ID = $3
			),
			@userId = (
				SELECT
					ID
				FROM
					USERS
				WHERE 
					ID = $4
			)
		IF @deviceId IS NULL
			BEGIN
				SELECT
					ERROR_ID = -7
				RETURN
			END
		IF @userId IS NULL
			BEGIN
				SELECT
					ERROR_ID = -8
				RETURN
			END
		IF @userDeviceExists IS NOT NULL
			BEGIN
				SELECT
					ERROR_ID = -9
				RETURN
			END
		ELSE
			BEGIN
				INSERT INTO USER_DEVICES(USER_ID, DEVICE_ID)
				VALUES (@userId, @deviceId)
				RETURN
			END
		END`,
				userDevices.UserID, userDevices.DeviceID,
				userDevices.DeviceID,
				userDevices.UserID)

		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
		}
		if mr.Status == -7 {
			iot.handleError(w, r, mr.Status, "device with given ID does not exist")
		} else if mr.Status == -8 {
			iot.handleError(w, r, mr.Status, "user with given ID does not exist")
		} else if mr.Status == -9 {
			iot.handleError(w, r, mr.Status, "user already has control over this device")
		} else {
			if sqlErr != nil {
				iot.handleError(w, r, mr.Status, sqlErr.Error())
			} else {
				fmt.Fprintf(w, "device: %+v", userDevices)
			}
		}
	}
}

func (iot *IOT) insertTemperatureInDB(deviceId int) {
	_, sqlErr := iot.database.Conn.Exec(`
	BEGIN
		declare @deviceId int
	SELECT 
		@deviceId = 
			(
				SELECT
					d.ID
				FROM
					DEVICES as d
				LEFT JOIN ROOMS AS r on d.ROOM_ID = r.ID
				LEFT JOIN
					FLOORS as f ON r.FLOOR_ID = f.ID
				LEFT JOIN
					BUILDINGS as b ON f.BUILDING_ID = b.ID
				WHERE 
					d.ID = $1
			)
		IF @deviceId IS NOT NULL
			BEGIN
				INSERT INTO	TEMPERATURES
					(DEVICE_ID, MEASUREMENT_UNIT_ID, IN_TEMPERATURE)
				VALUES (@deviceId, $2, $3)
			RETURN
		END
		ELSE
			BEGIN
				SELECT ID = -10
			RETURN
		END
	END`,
		deviceId,
		1, iot.data.Temperature)
	if sqlErr != nil {
		err := fmt.Sprint("Error ID:", -10, "\nError Message:", sqlErr.Error())
		iot.log.Error(err)
	}
}

func (iot *IOT) insertHumidityInDB(deviceId int) {
	_, sqlErr := iot.database.Conn.Exec(`
	BEGIN
		declare @deviceId int
	SELECT 
		@deviceId = 
			(
				SELECT
					d.ID
				FROM
					DEVICES as d
				LEFT JOIN ROOMS AS r on d.ROOM_ID = r.ID
				LEFT JOIN
					FLOORS as f ON r.FLOOR_ID = f.ID
				LEFT JOIN
					BUILDINGS as b ON f.BUILDING_ID = b.ID
				WHERE 
					d.ID = $1
			)
		IF @deviceId IS NOT NULL
			BEGIN
				INSERT INTO	HUMIDITIES
					(DEVICE_ID, HUMIDITY)
				VALUES (@deviceId, $2)
			RETURN
		END
		ELSE
			BEGIN
				SELECT ID = -11
			RETURN
		END
	END`,
		deviceId,
		iot.data.Humidity)
	if sqlErr != nil {
		err := fmt.Sprint("Error ID:", -11, "\nError Message:", sqlErr.Error())
		iot.log.Error(err)
	}
}

func (iot *IOT) getUserBuildingsAndDevices(w http.ResponseWriter, r *http.Request) {
	var userDevices user.User

	decErr := jsondecodebody.DecodeJSONBody(w, r, &userDevices)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			http.Error(w, mr.Msg, mr.Status)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	}
	fmt.Println("user", &userDevices)

	rows, err :=
		iot.database.Conn.Query(`
		BEGIN
		declare @userId int
		SELECT
			@userId = (
				SELECT TOP 1
					USER_ID
				FROM
					USER_DEVICES
				WHERE
					USER_ID = $1
	
			)
		IF @userId IS NOT NULL
			BEGIN
				SELECT
					b.ID as BUILDING_ID,
					b.NAME,
					b.CITY,
					(SELECT CONCAT(b.STREET, ' ', b.STREET_NUMBER)) AS STREET,
					f.ID as FLOOR_ID,
					f.FLOOR,
					r.ID as ROOM_ID,
					r.ROOM_NAME,
					ud.DEVICE_ID,
					d.IS_CONTROL_DEVICE,
				    d.IS_CONTROL_ON
				FROM
					USER_DEVICES as ud
				LEFT JOIN
					DEVICES as d ON ud.DEVICE_ID = d.ID
				LEFT JOIN
					ROOMS as r ON d.ROOM_ID = r.ID
				LEFT JOIN
					FLOORS as f ON r.FLOOR_ID = f.ID
				LEFT JOIN
					BUILDINGS as b ON f.BUILDING_ID = b.ID
				WHERE
					ud.USER_ID = @userId AND
					d.IS_DEVICE_ACTIVE = 1
				ORDER BY 
					b.ID ASC,
					f.ID,
					r.ID,
					d.ID
				RETURN
			END
		ELSE
			BEGIN
				SELECT USER_ID = -12
			RETURN
		END
	END`,
			userDevices.Id)
	if err != nil {
		iot.handleError(w, r, -12, err.Error())
	}

	var userBuildingDevices dbstructs.UserBuildingsAndDevices = dbstructs.UserBuildingsAndDevices{}
	for rows.Next() {
		var building dbstructs.Buildings = dbstructs.Buildings{}
		var floor dbstructs.Floors = dbstructs.Floors{}
		var room dbstructs.Rooms = dbstructs.Rooms{}
		var device dbstructs.Devices = dbstructs.Devices{}
		rows.Scan(
			&building.ID, &building.Name, &building.City, &building.Street,
			&floor.ID, &floor.Floor,
			&room.ID, &room.RoomName,
			&device.ID, &device.IsControlDevice, &device.IsControlOn)

		if len(userBuildingDevices.Buildings) == 0 {
			building.Floors = append(building.Floors, floor)
			building.Floors[len(building.Floors)-1].Rooms = append(building.Floors[len(building.Floors)-1].Rooms, room)
			building.Floors[len(building.Floors)-1].Rooms[len(building.Floors[len(building.Floors)-1].Rooms)-1].Devices = append(building.Floors[len(building.Floors)-1].Rooms[len(building.Floors[len(building.Floors)-1].Rooms)-1].Devices, device)
			userBuildingDevices.Buildings = append(userBuildingDevices.Buildings, building)
		} else {

			var buildingExists = userBuildingDevices.FindBuilding(building.ID)
			var floorExists = userBuildingDevices.FindFloor(floor.ID)
			var roomExists = userBuildingDevices.FindRoom(room.ID)
			var deviceExists = userBuildingDevices.FindDevice(device.ID)

			var buildingIndex = buildingExists[0]
			var floorIndex = floorExists[0]
			var roomIndex = roomExists[0]

			if uint(buildingExists[1]) == building.ID {

				if uint(floorExists[1]) == floor.ID {

					if uint(roomExists[1]) == room.ID {

						if uint(deviceExists[1]) != device.ID {
							userBuildingDevices.Buildings[buildingIndex].Floors[floorIndex].Rooms[roomIndex].Devices = append(userBuildingDevices.Buildings[buildingIndex].Floors[floorIndex].Rooms[roomIndex].Devices, device)
						}
					} else {
						room.Devices = append(room.Devices, device)
						userBuildingDevices.Buildings[buildingIndex].Floors[floorIndex].Rooms = append(userBuildingDevices.Buildings[buildingIndex].Floors[floorIndex].Rooms, room)
					}
				} else {
					room.Devices = append(room.Devices, device)
					floor.Rooms = append(floor.Rooms, room)
					userBuildingDevices.Buildings[buildingIndex].Floors = append(userBuildingDevices.Buildings[buildingIndex].Floors, floor)
					sort.Slice(userBuildingDevices.Buildings[buildingIndex].Floors, func(i, j int) bool {
						return userBuildingDevices.Buildings[buildingIndex].Floors[i].Floor < userBuildingDevices.Buildings[buildingIndex].Floors[j].Floor
					})
				}
			} else {
				room.Devices = append(room.Devices, device)
				floor.Rooms = append(floor.Rooms, room)
				building.Floors = append(building.Floors, floor)
				userBuildingDevices.Buildings = append(userBuildingDevices.Buildings, building)
			}
		}
	}

	mr := &iot.mr
	if userBuildingDevices.Buildings[0].ID == 0 {
		mr.Status = -12
		iot.handleError(w, r, mr.Status, "User does not control a device")
	} else {
		jsonData, _ := json.Marshal(userBuildingDevices)
		fmt.Fprintf(w, "%s", jsonData)
	}
}

func (iot *IOT) getDeviceData(w http.ResponseWriter, r *http.Request) {
	var device dbstructs.Devices

	decErr := jsondecodebody.DecodeJSONBody(w, r, &device)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			iot.handleError(w, r, mr.Status, mr.Msg)
		} else {
			iot.log.Error(decErr.Error())
			iot.handleError(w, r, mr.Status, decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		fmt.Println("device", &device)

		rows, err :=
			iot.database.Conn.Query(`
		BEGIN
		declare @deviceId int
		SELECT
			@deviceId = 
			(
				SELECT TOP 1
					t.ID
				FROM
					TEMPERATURES AS t
				INNER JOIN 
					HUMIDITIES AS h
				ON 
					t.DEVICE_ID = h.DEVICE_ID
				WHERE 
					t.DEVICE_ID = $1
			)
			IF @deviceId IS NOT NULL
				BEGIN
					SELECT TOP 1
						t.IN_TEMPERATURE,
						mu.UNIT_DESCRIPTION,
						h.HUMIDITY
					FROM
						TEMPERATURES as t
					LEFT JOIn
						MEASUREMENT_UNITS as mu
					ON
						t.MEASUREMENT_UNIT_ID = mu.ID
					INNER JOIN
						HUMIDITIES as h
					ON
						t.DEVICE_ID = h.DEVICE_ID
					WHERE
						t.DEVICE_ID = $2
					ORDER BY
						t.ID DESC
					RETURN
				END
			ELSE
				BEGIN
					SELECT ID = -14
				RETURN 
			END
		END`,
				device.ID,
				device.ID)

		if err != nil {
			iot.handleError(w, r, -13, err.Error())
		}

		var deviceData dbstructs.DeviceReadData
		for rows.Next() {
			rows.Scan(&deviceData.Temperature, &deviceData.TempUnit, &deviceData.Humidity)
		}
		mr := &iot.mr
		if deviceData.Temperature == 0 {
			mr.Status = -13
		}

		if mr.Status != 0 {
			iot.handleError(w, r, mr.Status, "Data for this device does not exist, please check the device")
		} else {
			jsonData, _ := json.Marshal(deviceData)
			fmt.Fprintf(w, "%s", jsonData)
		}
	}
}

func (iot *IOT) setDesiredTemperatureToDevice(w http.ResponseWriter, r *http.Request) {

	var heatingHistory dbstructs.HeatingHistory
	decErr := jsondecodebody.DecodeJSONBody(w, r, &heatingHistory)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			iot.handleError(w, r, mr.Status, mr.Msg)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		fmt.Println("heatingHistory", &heatingHistory)

		rows, err :=
			iot.database.Conn.Query(`
			BEGIN
			declare @deviceId int
			declare @measurementUnitId int
			declare @currentTemp float
			SELECT
				@deviceId =
				(
					SELECT TOP 1
						ud.ID
					FROM
						USER_DEVICES AS  ud
					WHERE
						ud.DEVICE_ID = $1
				),
				@currentTemp =
				(
					SELECT TOP 1
						t.IN_TEMPERATURE
					FROM
						TEMPERATURES AS t
					WHERE
						t.DEVICE_ID = $2
				),
				@measurementUnitId =
				(
					SELECT TOP 1
						us.ID
					FROM
						USERS_SETTINGS AS  us
					WHERE
						us.USER_ID = $3
				)
				IF @deviceId IS NOT NULL AND @measurementUnitId IS NOT NULL AND @currentTemp IS NOT NULL
					BEGIN
						INSERT INTO
							HEATING_HISTORY (USER_ID, DEVICE_ID, MEASUREMENT_UNIT_ID, DATE, TIME_FROM, CURRENT_TEMPERATURE, DESIRED_TEMPERATURE)
						VALUES
							($4, @deviceId, @measurementUnitId, (SELECT CAST (GETDATE() AS date)), (SELECT CAST (GETDATE() AS time(0))), @currentTemp, $5)
						RETURN
					END
				ELSE
					BEGIN
						SELECT ID = -15
					RETURN
				END
			END`,
				heatingHistory.DeviceID,
				heatingHistory.DeviceID,
				heatingHistory.UserId,
				heatingHistory.UserId,
				heatingHistory.DesiredTemperature)

		if err != nil {
			iot.handleError(w, r, -15, err.Error())
		}

		mr := &iot.mr
		for rows.Next() {
			rows.Scan(&mr.Status)
		}

		if mr.Status != 0 {
			iot.handleError(w, r, mr.Status, "Something went wrong, please try again")
		} else {
			jsonData, _ := json.Marshal(heatingHistory)
			fmt.Fprintf(w, "%s", jsonData)
		}
	}
}

func (iot *IOT) getHeatingHistory(w http.ResponseWriter, r *http.Request) {

	var user user.User
	decErr := jsondecodebody.DecodeJSONBody(w, r, &user)
	if decErr != nil {
		var mr *jsondecodebody.MalformedRequest
		if errors.As(decErr, &mr) {
			iot.handleError(w, r, mr.Status, mr.Msg)
		} else {
			iot.log.Error(decErr.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	} else {
		fmt.Println("heatingHistory", &user)

		rows, err :=
			iot.database.Conn.Query(`
			BEGIN
				declare @userId int
			SELECT
				@userId = (
					SELECT TOP 1
						USER_ID
					FROM
						USER_DEVICES
					WHERE
						USER_ID = $1
				)
			IF @userId IS NOT NULL
				BEGIN
					SELECT
						b.NAME,
						b.CITY,
						(SELECT CONCAT(b.STREET, ' ', b.STREET_NUMBER)) AS STREET,
						f.FLOOR,
						r.ROOM_NAME,
						ud.DEVICE_ID,
						hh.DATE,
						hh.TIME_FROM,
						hh.TIME_TO,
						hh.CURRENT_TEMPERATURE,
						hh.DESIRED_TEMPERATURE,
						mu.UNIT_DESCRIPTION
					FROM
						USER_DEVICES as ud
					LEFT JOIN
						DEVICES as d ON ud.DEVICE_ID = d.ID
					LEFT JOIN
						ROOMS as r ON d.ROOM_ID = r.ID
					LEFT JOIN
						FLOORS as f ON r.FLOOR_ID = f.ID
					LEFT JOIN
						BUILDINGS as b ON f.BUILDING_ID = b.ID
					RIGHT JOIN
						HEATING_HISTORY as hh ON ud.DEVICE_ID = hh.DEVICE_ID
					LEFT JOIN 
						MEASUREMENT_UNITS as mu ON hh.MEASUREMENT_UNIT_ID = mu.ID
					WHERE
						ud.USER_ID = @userId AND
						d.IS_DEVICE_ACTIVE = 1
					ORDER BY 
						b.ID ASC,
						f.ID,
						r.ID,
						d.ID
					RETURN
			END
			ELSE
				BEGIN
					SELECT USER_ID = -16
				RETURN
			END
			END;`,
				user.Id)

		if err != nil {
			iot.handleError(w, r, -15, err.Error())
		}

		var userDevicesHeatingHistories = dbstructs.UserDevicesHeatingHistories{}
		for rows.Next() {
			var userDevicesHeatingHistory dbstructs.UserDevicesHeatingHistory
			rows.Scan(&userDevicesHeatingHistory.BuildingName, &userDevicesHeatingHistory.City, &userDevicesHeatingHistory.Street, &userDevicesHeatingHistory.Floor, &userDevicesHeatingHistory.RoomName, &userDevicesHeatingHistory.DeviceID, &userDevicesHeatingHistory.Date, &userDevicesHeatingHistory.TimeFrom, &userDevicesHeatingHistory.TimeTo, &userDevicesHeatingHistory.CurrentTemperature, &userDevicesHeatingHistory.DesiredTemp, &userDevicesHeatingHistory.Unit)
			userDevicesHeatingHistories.UserDevicesHeatingHistory = append(userDevicesHeatingHistories.UserDevicesHeatingHistory, userDevicesHeatingHistory)
		}

		if len(userDevicesHeatingHistories.UserDevicesHeatingHistory) == 0 {
			var mr jsondecodebody.MalformedRequest
			mr.Status = -16
			iot.handleError(w, r, mr.Status, "No heating history")
		} else {
			jsonData, _ := json.Marshal(userDevicesHeatingHistories)
			fmt.Fprintf(w, "%s", jsonData)
		}
	}
}
