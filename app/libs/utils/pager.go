package utils

import (
    "bytes"
    "fmt"
    "math"
    "strings"
)

type Pager struct {
    page     int
    total    int
    pageSize int
    urlPath  string
    urlQuery string
    noPath   bool
}

func NewPager(page, total, pageSize int, url string, noPath ...bool) *Pager {
    p := &Pager{
        page:page,
        total:total,
        pageSize:pageSize,
    }
    arr := strings.Split(url, "?")
    p.urlPath = arr[0]
    if len(arr) > 1 {
        p.urlQuery = "?" + arr[1]
    } else {
        p.urlQuery = ""
    }

    if len(noPath) > 0 {
        p.noPath = noPath[0]
    } else {
        p.noPath = false
    }

    return p
}

func (p *Pager) url(page int) string {
    if p.noPath {
        //不使用目录形式
        if p.urlQuery != "" {
            return fmt.Sprintf("%s%s&page=%d", p.urlPath, p.urlQuery, page)
        } else {
            return fmt.Sprintf("%s?page=%d", p.urlPath, page)
        }
    } else {
        return fmt.Sprintf("%s/page/%d%s", p.urlPath, page, p.urlQuery)
    }
}

func (p *Pager) ToString() string {
    if p.total <= p.pageSize {
        return ""
    }

    var (
        buf bytes.Buffer
        from, to, totalPage int
        offset = 5
        linkNum = 10
    )

    totalPage = int(math.Ceil(float64(p.total) / float64(p.pageSize)))

    if totalPage < linkNum {
        from = 1
        to = totalPage
    } else {
        from = p.page - offset
        to = from + linkNum
        if from < 1 {
            from = 1
            to = from + linkNum - 1
        } else if to > totalPage {
            to = totalPage
            from = totalPage - linkNum + 1
        }
    }

    buf.WriteString("<ul class=\"pagination\">")
    if p.page > 1 {
        buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">&laquo;</a></li>", p.url(p.page - 1)))
    } else {
        buf.WriteString("<li class=\"disabled\"><span>&laquo;</span></li>")
    }

    if p.page > linkNum {
        buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">1...</a></li>", p.url(1)))
    }

    for i := from; i <= to; i++ {
        if i == p.page {
            buf.WriteString(fmt.Sprintf("<li class=\"active\"><span>%d</span></li>", i))
        } else {
            buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">%d</a></li>", p.url(i), i))
        }
    }

    if totalPage > to {
        buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">...%d</a></li>", p.url(totalPage), totalPage))
    }

    if p.page < totalPage {
        buf.WriteString(fmt.Sprintf("<li><a href=\"%s\">&raquo;</a></li>", p.url(p.page + 1)))
    } else {
        buf.WriteString(fmt.Sprintf("<li class=\"disabled\"><span>&raquo;</span></li>"))
    }
    buf.WriteString("</ul>")

    return buf.String()
}
