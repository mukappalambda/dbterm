package pages

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kevinliao852/dbterm/pkg/models"
	log "github.com/sirupsen/logrus"
)

type QueryPage struct {
	DbInput    textinput.Model
	DataTable  table.Model
	selectData string
	DB         *sql.DB
	queryStr   string
}

func NewQueryPage() QueryPage {
	return QueryPage{
		DbInput:    models.DBSQLQueryInput(),
		DataTable:  models.DBSelectTable(),
		selectData: "",
		DB:         nil,
		queryStr:   "",
	}
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.ThickBorder()).
	BorderForeground(lipgloss.Color("240")).Width(120)

var _ Pager = &QueryPage{}

var _ tea.Model = &QueryPage{}

func (q *QueryPage) Init() tea.Cmd {
	return textinput.Blink
}

func (q *QueryPage) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return q, tea.Quit
		case tea.KeyEnter:
			log.Println("Enter pressed")

			log.Println("DB:", q.DB, &q.DB)

			if q.DB == nil {
				log.Println("here")
				break
			}

			if q.DbInput.Value() == "" {
				q.selectData = "Please enter the SQL code"
				break
			}

			if q.DB.Ping() != nil {
				q.selectData = "DB is not connected"
				break
			}

			q.queryStr = q.DbInput.Value()
			q.selectData = "Querying the database"
			q.readAndQuery()

		}
	}

	q.DbInput, cmd = q.DbInput.Update(msg)

	return q, cmd
}

func (q *QueryPage) View() string {

	view := q.DbInput.View()

	return fmt.Sprintf("Select the database\n\n%s\n\n%s\n%s",
		view,
		q.selectData,
		baseStyle.Render(q.DataTable.View()),
	)
}

func (q *QueryPage) getPageName() string {
	return "queryPage"
}

func (q *QueryPage) readAndQuery() {
	if q.queryStr == "" {
		q.selectData = "Please enter the query"
		return
	}

	rows, err := q.DB.Query(q.queryStr)

	if err != nil {
		q.selectData = "Error executing the query\n" + err.Error()
		return
	}

	tableColumn := []table.Column{}
	tableRowList := []table.Row{}

	types, _ := rows.ColumnTypes()

	maxLength := 0

	for rows.Next() {

		row := make([]interface{}, 0)

		for range types {
			row = append(row, new(interface{}))
		}

		err := rows.Scan(row...)

		if err != nil {
			q.selectData = "Error scanning the row\n" + err.Error()
			return
		}

		var tableRow table.Row

		for _, fields := range row {
			pField := fields.(*interface{})
			strField := fmt.Sprintf("%s", *pField)
			maxLength = int(math.Max(float64(maxLength), float64(len(strField))))
			tableRow = append(tableRow, strField)
		}

		tableRowList = append(tableRowList, tableRow)
	}

	for _, col := range types {
		tableColumn = append(tableColumn, table.Column{
			Title: col.Name(),
			Width: maxLength,
		})
	}

	// make sure to set column first!
	if len(tableColumn) == 0 {
		var c []table.Column
		c = append(c, table.Column{Title: "Message", Width: 16})
		var r []table.Row
		r = append(r, table.Row{"No Rows Returned"})
		q.DataTable.SetColumns(c)
		q.DataTable.SetRows(r)
		return
	}

	q.DataTable.SetColumns(tableColumn)
	q.DataTable.SetRows(tableRowList)
}
