package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"time"
)

type PoolService interface {
	GetVersion() (VersionResponse, error)
	GetAvailableSets() (*AvailableSetsMap, error)
	GetLatestSet() (LatestSetResponse, error)

	CheckCubeList(list []string) ([]string, error)

	// Keep the pool service to determine what we want ?
	MakePool() ([]CardPool, error)

	// specific pools
	//MakeChaosPool(request ChaosRequest) []CardPool
	//MakeCubePool(request CubeRequest) []CardPool
	//MakeRegularPool(request RegularRequest) []CardPool
}

func NewPoolService(poolURL string) PoolService {
	c := cache.New(60*time.Minute, 10*time.Minute)
	return &defaultPoolService{poolURL: poolURL, cache: c}
}

type defaultPoolService struct {
	poolURL string
	cache   *cache.Cache
}

func (d *defaultPoolService) GetLatestSet() (latestSet LatestSetResponse, err error) {
	if cachedSet, ok := d.cache.Get("latestSet"); ok {
		return cachedSet.(LatestSetResponse), err
	}

	response, err := http.Get(fmt.Sprintf("%sets/latest", d.poolURL))
	if err != nil {
		return
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, &latestSet)
	if err == nil {
		d.cache.SetDefault("latestSet", latestSet)
	}

	return
}

func (d *defaultPoolService) GetVersion() (version VersionResponse, err error) {
	if cachedVersion, ok := d.cache.Get("version"); ok {
		return cachedVersion.(VersionResponse), err
	}

	response, err := http.Get(fmt.Sprintf("%sabout", d.poolURL))
	if err != nil {
		return
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, &version)
	if err == nil {
		d.cache.SetDefault("version", version)
	}
	return
}

func (d *defaultPoolService) GetAvailableSets() (sets *AvailableSetsMap, err error) {
	if cachedSetMap, ok := d.cache.Get("sets"); ok {
		return cachedSetMap.(*AvailableSetsMap), err
	}

	response, err := http.Get(fmt.Sprintf("%ssets", d.poolURL))
	if err != nil {
		return
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}

	err = json.Unmarshal(responseData, sets)
	if err == nil {
		d.cache.SetDefault("sets", sets)
	}

	return
}

func (d *defaultPoolService) CheckCubeList(list []string) ([]string, error) {
	request := CubeListRequest{list}
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("%scubelist", d.poolURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK {
		return []string{}, nil
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err == nil {
		return nil, err
	}

	var errorResponse CubeListErrorResponse
	err = json.Unmarshal(responseData, &errorResponse)

	return errorResponse.Error, err
}

func (d *defaultPoolService) MakePool() (pool []CardPool, err error) {
	panic("implement me")
}
