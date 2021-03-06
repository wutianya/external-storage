package testing

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/flavors"
	"github.com/gophercloud/gophercloud/pagination"
	th "github.com/gophercloud/gophercloud/testhelper"
	fake "github.com/gophercloud/gophercloud/testhelper/client"
)

const tokenID = "blerb"

func TestListFlavors(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors/detail", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		r.ParseForm()
		marker := r.Form.Get("marker")
		switch marker {
		case "":
			fmt.Fprintf(w, `
					{
						"flavors": [
							{
								"id": "1",
								"name": "m1.tiny",
								"vcpus": 1,
								"disk": 1,
								"ram": 512,
								"swap":"",
								"is_public": true
							},
							{
								"id": "2",
								"name": "m1.small",
								"vcpus": 1,
								"disk": 20,
								"ram": 2048,
								"swap": 1000,
								"is_public": true
							},
							{
								"id": "3",
								"name": "m1.medium",
								"vcpus": 2,
								"disk": 40,
								"ram": 4096,
								"swap": 1000,
								"is_public": false
							}
						],
						"flavors_links": [
							{
								"href": "%s/flavors/detail?marker=2",
								"rel": "next"
							}
						]
					}
				`, th.Server.URL)
		case "2":
			fmt.Fprintf(w, `{ "flavors": [] }`)
		default:
			t.Fatalf("Unexpected marker: [%s]", marker)
		}
	})

	pages := 0
	// Get public and private flavors
	err := flavors.ListDetail(fake.ServiceClient(), nil).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		expected := []flavors.Flavor{
			{ID: "1", Name: "m1.tiny", VCPUs: 1, Disk: 1, RAM: 512, Swap: 0, IsPublic: true},
			{ID: "2", Name: "m1.small", VCPUs: 1, Disk: 20, RAM: 2048, Swap: 1000, IsPublic: true},
			{ID: "3", Name: "m1.medium", VCPUs: 2, Disk: 40, RAM: 4096, Swap: 1000, IsPublic: false},
		}

		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("Expected %#v, but was %#v", expected, actual)
		}

		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if pages != 1 {
		t.Errorf("Expected one page, got %d", pages)
	}
}

func TestGetFlavor(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors/12345", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
			{
				"flavor": {
					"id": "1",
					"name": "m1.tiny",
					"disk": 1,
					"ram": 512,
					"vcpus": 1,
					"rxtx_factor": 1,
					"swap": ""
				}
			}
		`)
	})

	actual, err := flavors.Get(fake.ServiceClient(), "12345").Extract()
	if err != nil {
		t.Fatalf("Unable to get flavor: %v", err)
	}

	expected := &flavors.Flavor{
		ID:         "1",
		Name:       "m1.tiny",
		Disk:       1,
		RAM:        512,
		VCPUs:      1,
		RxTxFactor: 1,
		Swap:       0,
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, but was %#v", expected, actual)
	}
}

func TestCreateFlavor(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
			{
				"flavor": {
					"id": "1",
					"name": "m1.tiny",
					"disk": 1,
					"ram": 512,
					"vcpus": 1,
					"rxtx_factor": 1,
					"swap": ""
				}
			}
		`)
	})

	disk := 1
	opts := &flavors.CreateOpts{
		ID:         "1",
		Name:       "m1.tiny",
		Disk:       &disk,
		RAM:        512,
		VCPUs:      1,
		RxTxFactor: 1.0,
	}
	actual, err := flavors.Create(fake.ServiceClient(), opts).Extract()
	if err != nil {
		t.Fatalf("Unable to create flavor: %v", err)
	}

	expected := &flavors.Flavor{
		ID:         "1",
		Name:       "m1.tiny",
		Disk:       1,
		RAM:        512,
		VCPUs:      1,
		RxTxFactor: 1,
		Swap:       0,
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, but was %#v", expected, actual)
	}
}

func TestDeleteFlavor(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors/12345678", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "DELETE")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)

		w.WriteHeader(http.StatusAccepted)
	})

	res := flavors.Delete(fake.ServiceClient(), "12345678")
	th.AssertNoErr(t, res.Err)
}

func TestFlavorAccessesList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/flavors/12345678/os-flavor-access", func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", fake.TokenID)
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, `
			{
			  "flavor_access": [
			    {
			      "flavor_id": "12345678",
			      "tenant_id": "2f954bcf047c4ee9b09a37d49ae6db54"
			    }
			  ]
			}
		`)
	})

	expected := []flavors.FlavorAccess{
		flavors.FlavorAccess{
			FlavorID: "12345678",
			TenantID: "2f954bcf047c4ee9b09a37d49ae6db54",
		},
	}

	allPages, err := flavors.ListAccesses(fake.ServiceClient(), "12345678").AllPages()
	th.AssertNoErr(t, err)

	actual, err := flavors.ExtractAccesses(allPages)
	th.AssertNoErr(t, err)

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, but was %#v", expected, actual)
	}
}
