/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package module

/*
This file deals with the interaction of instantiation flow and etcd.
It contains methods for creating appContext, saving cluster and resource details to etcd.

*/
import (
	"encoding/json"
	"io/ioutil"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	gpic "github.com/onap/multicloud-k8s/src/orchestrator/pkg/gpic"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/resourcestatus"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils/helm"
	pkgerrors "github.com/pkg/errors"
)

// resource consists of name of reource
type resource struct {
	name        string
	filecontent string
}

type contextForCompositeApp struct {
	context            appcontext.AppContext
	ctxval             interface{}
	compositeAppHandle interface{}
}

// makeAppContext creates an appContext for a compositeApp and returns the output as contextForCompositeApp
func makeAppContextForCompositeApp(p, ca, v, rName string) (contextForCompositeApp, error) {
	context := appcontext.AppContext{}
	ctxval, err := context.InitAppContext()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}
	compositeHandle, err := context.CreateCompositeApp()
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error creating CompositeApp handle")
	}
	err = context.AddCompositeAppMeta(appcontext.CompositeAppMeta{Project: p, CompositeApp: ca, Version: v, Release: rName})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrap(err, "Error Adding CompositeAppMeta")
	}
	_, err = context.AddLevelValue(compositeHandle, "state", appcontext.AppContextState{State: appcontext.AppContextStateEnum.Init})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrapf(err, "Error setting initial AppContext State")
	}
	_, err = context.AddLevelValue(compositeHandle, "status", appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Pending})
	if err != nil {
		return contextForCompositeApp{}, pkgerrors.Wrapf(err, "Error setting initial AppContext Status")
	}

	m, err := context.GetCompositeAppMeta()

	log.Info(":: The meta data stored in the runtime context :: ", log.Fields{"Project": m.Project, "CompositeApp": m.CompositeApp, "Version": m.Version, "Release": m.Release})

	cca := contextForCompositeApp{context: context, ctxval: ctxval, compositeAppHandle: compositeHandle}

	return cca, nil

}

// getResources shall take in the sorted templates and output the resources
// which consists of name(name+kind) and filecontent
func getResources(st []helm.KubernetesResourceTemplate) ([]resource, error) {
	var resources []resource
	for _, t := range st {
		yamlStruct, err := utils.ExtractYamlParameters(t.FilePath)
		yamlFile, err := ioutil.ReadFile(t.FilePath)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Failed to get the resources..")
		}
		n := yamlStruct.Metadata.Name + SEPARATOR + yamlStruct.Kind
		// This might happen when the rendered file just has some comments inside, no real k8s object.
		if n == SEPARATOR {
			log.Info(":: Ignoring, Unable to render the template ::", log.Fields{"YAML PATH": t.FilePath})
			continue
		}

		resources = append(resources, resource{name: n, filecontent: string(yamlFile)})

		log.Info(":: Added resource into resource-order ::", log.Fields{"ResourceName": n})
	}
	return resources, nil
}

func addResourcesToCluster(ct appcontext.AppContext, ch interface{}, resources []resource) error {

	var resOrderInstr struct {
		Resorder []string `json:"resorder"`
	}

	var resDepInstr struct {
		Resdep map[string]string `json:"resdependency"`
	}
	resdep := make(map[string]string)

	for _, resource := range resources {
		resOrderInstr.Resorder = append(resOrderInstr.Resorder, resource.name)
		resdep[resource.name] = "go"
		_, err := ct.AddResource(ch, resource.name, resource.filecontent)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext after add resource failure ::", log.Fields{"Resource": resource.name, "Error": cleanuperr.Error})
			}
			return pkgerrors.Wrapf(err, "Error adding resource ::%s to AppContext", resource.name)
		}
		jresOrderInstr, _ := json.Marshal(resOrderInstr)
		resDepInstr.Resdep = resdep
		jresDepInstr, _ := json.Marshal(resDepInstr)
		_, err = ct.AddInstruction(ch, "resource", "order", string(jresOrderInstr))
		_, err = ct.AddInstruction(ch, "resource", "dependency", string(jresDepInstr))
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext after add instruction failure ::", log.Fields{"Resource": resource.name, "Error": cleanuperr.Error})
			}
			return pkgerrors.Wrapf(err, "Error adding instruction for resource ::%s to AppContext", resource.name)
		}
	}
	return nil
}

//addClustersToAppContext method shall add cluster details save into etcd
func addClustersToAppContext(l gpic.ClusterList, ct appcontext.AppContext, appHandle interface{}, resources []resource) error {
	mc := l.MandatoryClusters
	gc := l.ClusterGroups

	for _, c := range mc {
		p := c.ProviderName
		n := c.ClusterName
		clusterhandle, err := ct.AddCluster(appHandle, p+SEPARATOR+n)
		if err != nil {
			cleanuperr := ct.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Info(":: Error Cleaning up AppContext after add cluster failure ::", log.Fields{"cluster-provider": p, "cluster-name": n, "Error": cleanuperr.Error})
			}
			return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
		}

		err = addResourcesToCluster(ct, clusterhandle, resources)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s and name::%s) to AppContext", p, n)
		}
	}

	for _, eachGrp := range gc {
		oc := eachGrp.OptionalClusters
		gn := eachGrp.GroupNumber

		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName

			clusterhandle, err := ct.AddCluster(appHandle, p+SEPARATOR+n)

			if err != nil {
				cleanuperr := ct.DeleteCompositeApp()
				if cleanuperr != nil {
					log.Info(":: Error Cleaning up AppContext after add cluster failure ::", log.Fields{"cluster-provider": p, "cluster-name": n, "GroupName": gn, "Error": cleanuperr.Error})
				}
				return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
			}

			err = ct.AddClusterMetaGrp(clusterhandle, gn)
			if err != nil {
				cleanuperr := ct.DeleteCompositeApp()
				if cleanuperr != nil {
					log.Info(":: Error Cleaning up AppContext after add cluster failure ::", log.Fields{"cluster-provider": p, "cluster-name": n, "GroupName": gn, "Error": cleanuperr.Error})
				}
				return pkgerrors.Wrapf(err, "Error adding Cluster(provider::%s and name::%s) to AppContext", p, n)
			}

			err = addResourcesToCluster(ct, clusterhandle, resources)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error adding Resources to Cluster(provider::%s, name::%s and groupName:: %s) to AppContext", p, n, gn)
			}
		}
	}
	return nil
}

/*
verifyResources method is just to check if the resource handles are correctly saved.
*/
func verifyResources(l gpic.ClusterList, ct appcontext.AppContext, resources []resource, appName string) error {

	for _, cg := range l.ClusterGroups {
		gn := cg.GroupNumber
		oc := cg.OptionalClusters
		for _, eachCluster := range oc {
			p := eachCluster.ProviderName
			n := eachCluster.ClusterName
			cn := p + SEPARATOR + n

			for _, res := range resources {
				rh, err := ct.GetResourceHandle(appName, cn, res.name)
				if err != nil {
					return pkgerrors.Wrapf(err, "Error getting resoure handle for resource :: %s, app:: %s, cluster :: %s, groupName :: %s", appName, res.name, cn, gn)
				}
				log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": appName, "Cluster": cn, "Resource": res.name})
			}
		}
		grpMap, err := ct.GetClusterGroupMap(appName)
		if err != nil {
			return pkgerrors.Wrapf(err, "Error getting GetGroupMap for app:: %s, groupName :: %s", appName, gn)
		}
		log.Info(":: GetGroupMapReults ::", log.Fields{"GroupMap": grpMap})
	}

	for _, mc := range l.MandatoryClusters {
		p := mc.ProviderName
		n := mc.ClusterName
		cn := p + SEPARATOR + n
		for _, res := range resources {
			rh, err := ct.GetResourceHandle(appName, cn, res.name)
			if err != nil {
				return pkgerrors.Wrapf(err, "Error getting resoure handle for resource :: %s, app:: %s, cluster :: %s", appName, res.name, cn)
			}
			log.Info(":: GetResourceHandle ::", log.Fields{"ResourceHandler": rh, "appName": appName, "Cluster": cn, "Resource": res.name})
		}
	}
	return nil
}

// applyInitialResourceStatus sets the initial status of every resource in the appcontext to Pending
func applyInitialResourceStatus(ac appcontext.AppContext) error {
	status := resourcestatus.ResourceStatus{
		Status: resourcestatus.RsyncStatusEnum.Pending,
	}

	appsOrder, err := ac.GetAppInstruction("order")
	if err != nil {
		return err
	}
	var appList map[string][]string
	json.Unmarshal([]byte(appsOrder.(string)), &appList)

	for _, app := range appList["apporder"] {
		clusterNames, err := ac.GetClusterNames(app)
		if err != nil {
			return err
		}
		for k := 0; k < len(clusterNames); k++ {
			cluster := clusterNames[k]
			resorder, err := ac.GetResourceInstruction(app, cluster, "order")
			if err != nil {
				logutils.Error("Initial resource status resorder error", logutils.Fields{
					"error":   err,
					"app":     app,
					"cluster": cluster,
				})
				return err
			}
			var aov map[string][]string
			json.Unmarshal([]byte(resorder.(string)), &aov)
			for _, res := range aov["resorder"] {
				rh, err := ac.GetResourceHandle(app, cluster, res)
				if err != nil {
					return err
				}
				sh, err := ac.GetLevelHandle(rh, "status")
				if sh == nil {
					_, err = ac.AddLevelValue(rh, "status", status)
				} else {
					err = ac.UpdateStatusValue(sh, status)
				}
				if err != nil {
					logutils.Error("Set initial resource status error", logutils.Fields{
						"error":    err,
						"app":      app,
						"cluster":  cluster,
						"resource": res,
					})
					return err
				}
			}
		}
	}
	return nil
}
