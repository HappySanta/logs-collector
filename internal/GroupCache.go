package internal

import (
	"github.com/stels-cs/vk-api-tools"
	"log"
	"strconv"
	"strings"
)

var cache *UserPoll

type Group struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type UserPoll struct {
	poll   map[string]Group
	api    *VkApi.Api
	logger *log.Logger
}

func GetUserPoll(api *VkApi.Api, logger *log.Logger) *UserPoll {
	return &UserPoll{map[string]Group{}, api, logger}
}

func Get(groupIds []string) map[string]Group {
	if cache == nil {
		return make(map[string]Group)
	}
	return cache.Get(groupIds)
}

func InitCache(logger *log.Logger, token string) {
	t := VkApi.GetHttpTransport()
	cache = GetUserPoll(VkApi.CreateApi(token, "5.101", t, 2), logger)
}

func (up *UserPoll) Get(groupIds []string) map[string]Group {
	result := map[string]Group{}

	if len(groupIds) <= 0 {
		return result
	}

	var toRequest []string
	for _, v := range groupIds {
		if u, ok := up.poll[v]; ok == false {
			toRequest = append(toRequest, v)
			result[v] = Group{Id: strToInt(v), Name: "club" + v}
		} else {
			result[v] = u
		}
	}

	for len(toRequest) > 0 {
		if len(toRequest) > 500 {
			part := toRequest[:500]
			toRequest = toRequest[500:]
			result = AssignMap(result, up.req(part))
		} else {
			result = AssignMap(result, up.req(toRequest))
			toRequest = []string{}
		}
	}

	return result
}

func (up *UserPoll) req(toRequest []string) map[string]Group {
	result := map[string]Group{}
	users := make([]Group, 0, len(toRequest))
	err := up.api.Exec("groups.getById", VkApi.P{
		"group_ids": strings.Join(toRequest, ","),
	}, &users)
	if err == nil {
		for _, v := range users {
			result[intToString(v.Id)] = v
			up.poll[intToString(v.Id)] = v
		}
	} else {
		up.logger.Println(err)
	}
	return result
}

func intToString(id int) string {
	return strconv.Itoa(id)
}

func strToInt(str string) int {
	res, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return res
}

func AssignMap(origin, add map[string]Group) map[string]Group {
	for k, v := range add {
		origin[k] = v
	}
	return origin
}

func (up *UserPoll) Clear() {
	if len(up.poll) > 10000 {
		up.poll = map[string]Group{}
	}
}
