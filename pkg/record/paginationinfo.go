package record

import (
	"log"
	"strconv"
	"strings"
)

type PaginationInfo struct{}

const DEFAULT_PAGE_SIZE = 20

func (pi *PaginationInfo) HasPage(params map[string][]string) bool {
	_, exists := params["page"]
	return exists
}

func (pi *PaginationInfo) GetPageOffset(params map[string][]string) int {
	offset := 0
	pageSize := pi.getPageSize(params)
	if pages, exists := params["page"]; exists {
		for _, page := range pages {
			parts := strings.SplitN(page, ",", 2)
			parts_int, err := strconv.Atoi(parts[0])
			if err != nil {
				log.Printf("Error converting page parameter %v to int\n", parts[0])
				return offset
			}
			offset = (parts_int - 1) * pageSize
		}
	}
	return offset
}

func (pi *PaginationInfo) getPageSize(params map[string][]string) int {
	pageSize := DEFAULT_PAGE_SIZE
	if pages, exists := params["page"]; exists {
		for _, page := range pages {
			parts := strings.SplitN(page, ",", 2)
			if len(parts) > 1 {
				var err error
				pageSize, err = strconv.Atoi(parts[1])
				if err != nil {
					log.Printf("Error converting page parameter %v to int\n", parts[1])
					return DEFAULT_PAGE_SIZE
				}
			}
		}
	}
	return pageSize
}

func (pi *PaginationInfo) GetResultSize(params map[string][]string) int {
	numberOfRows := -1
	if sizes, exists := params["size"]; exists {
		for _, size := range sizes {
			var err error
			numberOfRows, err = strconv.Atoi(size)
			if err != nil {
				log.Printf("Error converting size parameter %v to int\n", size)
				return -1
			}
		}
	}
	return numberOfRows
}

func (pi *PaginationInfo) min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func (pi *PaginationInfo) GetPageLimit(params map[string][]string) int {
	pageLimit := -1
	if pi.HasPage(params) {
		pageLimit = pi.getPageSize(params)
	}
	resultSize := pi.GetResultSize(params)
	if resultSize >= 0 {
		if pageLimit >= 0 {
			pageLimit = pi.min(pageLimit, resultSize)
		} else {
			pageLimit = resultSize
		}
	}
	return pageLimit
}
