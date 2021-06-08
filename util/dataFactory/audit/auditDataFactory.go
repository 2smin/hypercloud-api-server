package audit

import (

	// pq "github.com/lib/pq"
	"context"
	"fmt"
	"os"

	pgx "github.com/jackc/pgx/v4"
	db "github.com/tmax-cloud/hypercloud-api-server/util/dataFactory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog/v2"
)

func Insert(items []audit.Event) {
	queryInsertTimeseriesData := `
	INSERT INTO audit (ID, USERNAME, USERAGENT , NAMESPACE , APIGROUP , APIVERSION , RESOURCE , NAME , STAGE , STAGETIMESTAMP , VERB, CODE , STATUS , REASON , MESSAGE ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);
	`

	/********************************************/
	/* Batch Insert into TimescaelDB            */
	/********************************************/
	//create batch
	batch := &pgx.Batch{}
	numInserts := len(items)
	//load insert statements into batch queue
	for _, event := range items {
		batch.Queue(queryInsertTimeseriesData, event.AuditID,
			event.User.Username,
			event.UserAgent,
			event.ObjectRef.Namespace,
			event.ObjectRef.APIGroup,
			event.ObjectRef.APIVersion,
			event.ObjectRef.Resource,
			event.ObjectRef.Name,
			event.Stage,
			event.StageTimestamp.Time,
			event.Verb,
			event.ResponseStatus.Code,
			event.ResponseStatus.Status,
			event.ResponseStatus.Reason,
			event.ResponseStatus.Message)
	}
	//send batch to connection pool
	br := db.Dbpool.SendBatch(context.TODO(), batch)
	//execute statements in batch queue
	for i := 0; i < numInserts; i++ {
		_, err := br.Exec()
		if err != nil {
			klog.Errorln(err)
			// os.Exit(1)
		}
	}
	// klog.Infoln("Successfully batch inserted data n")

	//Compare length of results slice to size of table
	klog.Infof("size of results: %d\n", numInserts)
	//check size of table for number of rows inserted
	// result of last SELECT statement
	var rowsInserted int
	err := br.QueryRow().Scan(&rowsInserted)
	klog.Infof("size of table: %d\n", rowsInserted)

	err = br.Close()
	if err != nil {
		klog.Errorf("Unable to closer batch %v\n", err)
	}

}

func Get(query string) (audit.EventList, int64) {
	// role check
	rows, err := db.Dbpool.Query(context.TODO(), query)
	if err != nil {
		klog.Errorf("Unable to execute query %v\n", err)
	}
	defer rows.Close()
	var row_count int64
	// Print rows returned and fill up results slice for later use
	var eventList audit.EventList
	for rows.Next() {
		event := audit.Event{
			ObjectRef:      &audit.ObjectReference{},
			ResponseStatus: &metav1.Status{},
		}
		err = rows.Scan(
			&event.AuditID,
			&event.User.Username,
			&event.UserAgent,
			&event.ObjectRef.Namespace,
			&event.ObjectRef.APIGroup,
			&event.ObjectRef.APIVersion,
			&event.ObjectRef.Resource,
			&event.ObjectRef.Name,
			&event.Stage,
			&event.StageTimestamp.Time,
			&event.Verb,
			&event.ResponseStatus.Code,
			&event.ResponseStatus.Status,
			&event.ResponseStatus.Reason,
			&event.ResponseStatus.Message,
			&row_count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to scan %v\n", err)
			os.Exit(1)
		}
		eventList.Items = append(eventList.Items, event)
	}
	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		klog.Errorf("rows Error: %v\n", rows.Err())
	}

	// use results here…
	eventList.Kind = "EventList"
	eventList.APIVersion = "audit.k8s.io/v1"

	return eventList, row_count
}
