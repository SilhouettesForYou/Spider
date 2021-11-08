package models

import (
	"github.com/monnand/goredis"
)

const (
	CONTENT_SUCCESS_SET    = "CONTENT_SUCCESS_SET"
	CONTENT_UNSUCCESS_SET  = "CONTENT_UNSUCCESS_SET"
	CONTENT_URL_VISIT_SET  = "CONTENT_URL_VISITED"
	MAGNET_SUCCESS_SET 	   = "MAGNET_SUCCESS_SET"
	MAGNET_UNSUCCESS_SET   = "MAGNET_UNSUCCESS_SET"
	MAGNET_URL_VISIT_SET   = "MAGNET_URL_VISIT_SET"
	MOVIE_HEAVEN_VISIT_SET = "MOVIE_HEAVEN_VISIT_SET"
	NEW_TIMES_VISIT_SET	   = "NEW_TIMES_VISIT_SET"
	STORE_FAILED_IMAGE_SET = "STORE_FAILED_IMAGE_SET"
)

var (
	client goredis.Client
)

func ConnectRedis(addr string) {
	client.Addr = addr
}

func AddToContentSuccess(url string) {
	client.Sadd(CONTENT_SUCCESS_SET, []byte(url))
}

func AddToContentUnsuccess(url string) {
	client.Sadd(CONTENT_UNSUCCESS_SET, []byte(url))
}

func AddToContentVisitedSet(url string) {
	client.Sadd(CONTENT_URL_VISIT_SET, []byte(url))
}

func AddToMagnetSuccess(url string) {
	client.Sadd(MAGNET_SUCCESS_SET, []byte(url))
}

func AddToMagnetUnsuccess(url string) {
	client.Sadd(MAGNET_UNSUCCESS_SET, []byte(url))
}

func AddToMagnetVisitedSet(url string) {
	client.Sadd(MAGNET_URL_VISIT_SET, []byte(url))
}

func AddToMovieHeavenVisitedSet(name string) {
	client.Sadd(MOVIE_HEAVEN_VISIT_SET, []byte(name))
}

func AddToNewTimesVisitedSet(url string) {
	client.Sadd(NEW_TIMES_VISIT_SET, []byte(url))
}

func AddToStoreFailedImageSet(src string) {
	client.Sadd(STORE_FAILED_IMAGE_SET, []byte(src))
}

func IsContentVisit(url string) bool {
	isVisit, err := client.Sismember(CONTENT_URL_VISIT_SET, []byte(url))
	if err != nil {
		return false
	}
	return isVisit
}

func IsMagnetVisit(url string) bool {
	isVisit, err := client.Sismember(MAGNET_URL_VISIT_SET, []byte(url))
	if err != nil {
		return false
	}
	return isVisit
}

func IsMovieHeavenVisit(url string) bool {
	isVisit, err := client.Sismember(MOVIE_HEAVEN_VISIT_SET, []byte(url))
	if err != nil {
		return false
	}
	return isVisit
}

func IsNewTimesVisit(url string) bool {
	isVisit, err := client.Sismember(NEW_TIMES_VISIT_SET, []byte(url))
	if err != nil {
		return false
	}
	return isVisit
}

func RemoveItemFromRedis(key, value string) {
	client.Srem(key, []byte(value))
}

func ClearContentSuccess() {
	for {
		size, _ := client.Scard(CONTENT_UNSUCCESS_SET)
		if size == 0 {
			break
		}
		client.Spop(CONTENT_UNSUCCESS_SET)
	} 
}

func ClearContentUnsuccess() {
	for {
		size, _ := client.Scard(CONTENT_UNSUCCESS_SET)
		if size == 0 {
			break
		}
		client.Spop(CONTENT_UNSUCCESS_SET)
	} 
}

func ClearContentVisited() {
	for {
		size, _ := client.Scard(CONTENT_URL_VISIT_SET)
		if size == 0 {
			break
		}
		client.Spop(CONTENT_URL_VISIT_SET)
	} 
}

func ClearMagnetSuccess() {
	for {
		size, _ := client.Scard(MAGNET_SUCCESS_SET)
		if size == 0 {
			break
		}
		client.Spop(MAGNET_SUCCESS_SET)
	} 
}

func ClearMagnetUnsuccess() {
	for {
		size, _ := client.Scard(MAGNET_UNSUCCESS_SET)
		if size == 0 {
			break
		}
		client.Spop(MAGNET_UNSUCCESS_SET)
	} 
}

func ClearMagnetVisited() {
	for {
		size, _ := client.Scard(MAGNET_URL_VISIT_SET)
		if size == 0 {
			break
		}
		client.Spop(MAGNET_URL_VISIT_SET)
	} 
}

func ClearHeavenVisited() {
	for {
		size, _ := client.Scard(MOVIE_HEAVEN_VISIT_SET)
		if size == 0 {
			break
		}
		client.Spop(MOVIE_HEAVEN_VISIT_SET)
	} 
}

func ClearNewTimesVisited() {
	for {
		size, _ := client.Scard(NEW_TIMES_VISIT_SET)
		if size == 0 {
			break
		}
		client.Spop(NEW_TIMES_VISIT_SET)
	} 
}

func ClearAll() {
	ClearContentSuccess()
	ClearContentUnsuccess()
	ClearContentVisited()
	ClearMagnetSuccess()
	ClearMagnetUnsuccess()
	ClearMagnetVisited()
	ClearHeavenVisited()
	ClearNewTimesVisited()
}