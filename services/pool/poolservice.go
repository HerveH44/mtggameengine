package pool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"mtggameengine/models"
	"net/http"
	"time"
)

type Service interface {
	GetVersion() (models.VersionResponse, error)
	GetAvailableSets() (*models.AvailableSetsMap, error)
	GetLatestSet() (models.LatestSetResponse, error)
	CheckCubeList(list []string) ([]string, error)

	// specific pools
	MakeChaosPool(request models.ChaosRequest) (models.Pool, error)
	MakeCubePool(request models.CubeRequest) (models.Pool, error)
	MakeRegularPool(request models.RegularRequest) (models.Pool, error)
}

func NewPoolService(poolURL string) Service {
	c := cache.New(60*time.Minute, 10*time.Minute)
	return &defaultPoolService{poolURL: poolURL, cache: c}
}

type defaultPoolService struct {
	poolURL string
	cache   *cache.Cache
}

func (d *defaultPoolService) GetLatestSet() (latestSet models.LatestSetResponse, err error) {
	if cachedSet, ok := d.cache.Get("latestSet"); ok {
		return cachedSet.(models.LatestSetResponse), err
	}

	response, err := http.Get(fmt.Sprintf("%ssets/latest", d.poolURL))
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

func (d *defaultPoolService) GetVersion() (version models.VersionResponse, err error) {
	if cachedVersion, ok := d.cache.Get("version"); ok {
		return cachedVersion.(models.VersionResponse), err
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

func (d *defaultPoolService) GetAvailableSets() (sets *models.AvailableSetsMap, err error) {
	setsMap := make(models.AvailableSetsMap)
	sets = &setsMap
	if cachedSetMap, ok := d.cache.Get("sets"); ok {
		return cachedSetMap.(*models.AvailableSetsMap), err
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
	request := models.CubeListRequest{Cubelist: list}
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
	if err != nil {
		return nil, err
	}

	var errorResponse models.CubeListErrorResponse
	err = json.Unmarshal(responseData, &errorResponse)

	return errorResponse.Error, err
}

func (d *defaultPoolService) MakeRegularPool(request models.RegularRequest) (pool models.Pool, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("%sregular", d.poolURL), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error while fetching regular pool")
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseData, &pool)

	return
}
func (d *defaultPoolService) MakeChaosPool(request models.ChaosRequest) (pool models.Pool, err error) {
	return makePool(fmt.Sprintf("%schaos", d.poolURL), request)
}

func (d *defaultPoolService) MakeCubePool(request models.CubeRequest) (models.Pool, error) {
	return makePool(fmt.Sprintf("%scube", d.poolURL), request)
}

func makePool(url string, request interface{}) (pool models.Pool, err error) {
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected error while fetching regular pool")
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(responseData, &pool)
	return
}
