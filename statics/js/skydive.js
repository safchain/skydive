/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

var hostImg = 'statics/img/host.png';
var switchImg = 'statics/img/switch.png';
var portImg = 'statics/img/port.png';
var intfImg = 'statics/img/intf.png';
var vethImg = 'statics/img/veth.png';
var nsImg = 'statics/img/ns.png';
var bridgeImg = 'statics/img/bridge.png';
var dockerImg = 'statics/img/docker.png';
var neutronImg = 'statics/img/openstack.png';
var minusImg = 'statics/img/minus-outline-16.png';
var plusImg = 'statics/img/plus-16.png';
var probeIndicatorImg = 'statics/img/media-record.png';
var trashImg = 'statics/img/trash.png';

var alerts = {};

var CurrentNodeDetails;
var NodeDetailsTmID;

var Group = function(ID, type) {
  this.ID = ID;
  this.Type = type;
};

var Node = function(ID) {
  this.ID = ID;
  this.Host = '';
  this.Metadata = {};
  this.Edges = {};
  this.Visible = true;
  this.Collapsed = false;
  this.Group = '';
};

Node.prototype.IsCaptureOn = function() {
  return "State.FlowCapture" in this.Metadata && this.Metadata["State.FlowCapture"] == "ON";
};

Node.prototype.IsCaptureAllowed = function() {
  var allowedTypes = ["device", "veth", "ovsbridge", "internal", "tun", "bridge"];
  return allowedTypes.indexOf(this.Metadata.Type) >= 0;
};

var Edge = function(ID) {
  this.ID = ID;
  this.Host = '';
  this.Parent = '';
  this.Child = '';
  this.Metadata = {};
  this.Visible = true;
};

var Graph = function(ID) {
  this.Nodes = {};
  this.Edges = {};
  this.Groups = {};
};

Graph.prototype.NewNode = function(ID, host) {
  var node = new Node(ID);
  node.Graph = this;
  node.Host = host;

  this.Nodes[ID] = node;

  return node;
};

Graph.prototype.GetNode = function(ID) {
  return this.Nodes[ID];
};

Graph.prototype.GetEdge = function(ID) {
  return this.Edges[ID];
};

Graph.prototype.NewEdge = function(ID, parent, child, host) {
  var edge = new Edge(ID);
  edge.Parent = parent;
  edge.Child = child;
  edge.Graph = this;
  edge.Host = host;

  this.Edges[ID] = edge;

  parent.Edges[ID] = edge;
  child.Edges[ID] = edge;

  return edge;
};

Graph.prototype.DelNode = function(node) {
  for (var i in node.Edges) {
    this.DelEdge(this.Edges[i]);
  }

  delete this.Nodes[node.ID];
};

Graph.prototype.DelEdge = function(edge) {
  delete edge.Parent.Edges[edge.ID];
  delete edge.Child.Edges[edge.ID];
  delete this.Edges[edge.ID];
};

Graph.prototype.InitFromSyncMessage = function(msg) {
  var g = msg.Obj;

  var i;
  for (i in g.Nodes) {
    var n = g.Nodes[i];

    var node = this.NewNode(n.ID);
    if ("Metadata" in n)
      node.Metadata = n.Metadata;
    node.Host = n.Host;
  }

  for (i in g.Edges) {
    var e = g.Edges[i];

    var parent = this.GetNode(e.Parent);
    var child = this.GetNode(e.Child);

    var edge = this.NewEdge(e.ID, parent, child);

    if ("Metadata" in e)
      edge.Metadata = e.Metadata;
    edge.Host = e.Host;
  }
};

var Layout = function(selector) {
  this.graph = new Graph();
  this.selector = selector;
  this.updatesocket = '';
  this.elements = {};

  this.width = $(selector).width() - 20;
  this.height = $(selector).height();

  this.svg = d3.select(selector).append("svg")
    .attr("width", this.width)
    .attr("height", this.height)
    .attr("y", 60)
    .attr('viewBox', -this.width/2 + ' ' + -this.height/2 + ' ' + this.width * 2 + ' ' + this.height * 2)
    .attr('preserveAspectRatio', 'xMidYMid meet')
    .call(d3.behavior.zoom().on("zoom", function() {
      _this.Rescale();
    }));

  var _this = this;
  this.force = d3.layout.force()
    .size([this.width, this.height])
    .charge(-600)
    .gravity(0.04)
    //.friction(0.5)
    .linkDistance(this.LinkDistance)
    .on("tick", function(e) {
      _this.Tick(e);
    });

  this.view = this.svg.append('g');

  this.drag = this.force.stop().drag()
    .on("dragstart", function(d) {
      d3.event.sourceEvent.stopPropagation();
    });

  this.groupsG = this.view.append("g")
    .attr("class", "groups")
    .on("click", function() {
      d3.event.preventDefault();
    });

  this.links = this.force.links();
  this.nodes = this.force.nodes();

  var linksG = this.view.append("g").attr("class", "links");
  this.link = linksG.selectAll(".link");

  var nodesG = this.view.append("g").attr("class", "nodes");
  this.node = nodesG.selectAll(".node");
};

Layout.prototype.LinkDistance = function(d, i) {
  if (d.source.Group == d.target.Group)
    return 70;

  // local to fabric
  if ((d.source.Metadata.Probe == "fabric" && !d.target.Metadata.Probe) ||
      (!d.source.Metadata.Probe && d.target.Metadata.Probe == "fabric")) {
    return 300;
  }
  return 120;
};

Layout.prototype.InitFromSyncMessage = function(msg) {
  this.graph.InitFromSyncMessage(msg);

  var ID;
  for (ID in this.graph.Nodes)
    this.AddNode(this.graph.Nodes[ID]);

  for (ID in this.graph.Edges)
    this.AddEdge(this.graph.Edges[ID]);
};

Layout.prototype.Clear = function() {
  var ID;

  for (ID in this.graph.Edges)
    this.DelEdge(this.graph.Edges[ID]);

  for (ID in this.graph.Nodes)
    this.DelNode(this.graph.Nodes[ID]);

  for (ID in this.graph.Edges)
    this.graph.DelEdge(this.graph.Edges[ID]);

  for (ID in this.graph.Nodes)
    this.graph.DelNode(this.graph.Nodes[ID]);
};

Layout.prototype.Rescale = function() {
  var trans = d3.event.translate;
  var scale = d3.event.scale;

  this.view.attr("transform", "translate(" + trans + ")" + " scale(" + scale + ")");
};

Layout.prototype.SetPosition = function(x, y) {
  this.view.attr("x", x).attr("y", y);
};

function ShowNodeFlows(node) {
  var query = "G.V('" + node.ID + "').Flows().Limit(5)";
  $.ajax({
    dataType: "json",
    url: '/api/topology',
    data: JSON.stringify({"GremlinQuery": query}),
    method: 'POST',
    success: function(data) {
      var packets = 0;
      var bytes = 0;
      for (var i in data) {
        for (var j in data[i].Statistics.Endpoints) {
          var endpoint = data[i].Statistics.Endpoints[j];
          if (endpoint.Type == "ETHERNET") {
            packets += endpoint.AB.Packets;
            bytes += endpoint.AB.Bytes;
          }
        }
      }

      $("#flow-packets").html(packets);
      $("#flow-bytes").html(bytes);

      var json = JSON.stringify(data);
      $("#flows").JSONView(json);
      $('#flows').JSONView('toggle', 10);
    }
  });
}

Layout.prototype.NodeDetails = function(node) {
  CurrentNodeDetails = node;
  $("#node-details").show();

  var json = JSON.stringify(node.Metadata);
  $("#metadata").JSONView(json);
  $("#node-id").html(node.ID);

  ShowNodeFlows(node);

  if (node.IsCaptureAllowed()) {
    if (node.IsCaptureOn()) {
      $("#add-capture").parent().css("cursor","not-allowed");
      $("#add-capture").attr("src", "statics/img/record_red.png").css("pointer-events","none");
    } else {
      $("#add-capture").parent().css("cursor","auto");
      $("#add-capture").attr("src", "statics/img/record.png").css({"cursor":"pointer", "pointer-events":"auto"});
    }
  } else {
    $("#add-capture").parent().css("cursor","not-allowed");
    $("#add-capture").attr("src", "statics/img/record.png").css("pointer-events","none");
  }
};

Layout.prototype.AddNode = function(node) {
  if (node.ID in this.elements)
    return;

  this.elements[node.ID] = node;
  this.nodes.push(node);

  this.Redraw();
};

Layout.prototype.UpdateNode = function(node, metadata) {
  node.Metadata = metadata;

  if (typeof CurrentNodeDetails != "undefined" && node.ID == CurrentNodeDetails.ID)
    this.NodeDetails(node);

  this.Redraw();
};

Layout.prototype.DelNode = function(node) {
  if (!(node.ID in this.elements))
    return;

  for (var i in this.nodes) {
    if (this.nodes[i].ID == node.ID) {
      this.nodes.splice(i, 1);
      break;
    }
  }
  delete this.elements[node.ID];

  this.Redraw();
};

Layout.prototype.AddEdge = function(edge) {
  if (edge.ID in this.elements)
    return;

  this.elements[edge.ID] = edge;

  // ignore layer 3 for now
  if (edge.Metadata.RelationType == "layer3")
    return;

  // specific to link to host
  var i, e;
  if (edge.Parent.Metadata.Type == "host") {
    // do not link ovsbridge to host
    if (edge.Child.Metadata.Type == "ovsbridge")
      return;

    // do not add link if having children
    var child = edge.Child;
    for (i in child.Edges) {
      e = child.Edges[i];
      if (e.Parent == child) {
        return;
      }
    }
  } else {
    // removing link to host if having children
    var parent = edge.Parent;
    for (i in parent.Edges) {
      e = parent.Edges[i];
      if (e.Child == parent && e.Parent.Metadata.Type == "host") {
        this.DelEdge(e);
        break;
      }
    }
  }

  this.links.push({source: edge.Parent, target: edge.Child, edge: edge});

  this.Redraw();
};

Layout.prototype.DelEdge = function(edge) {
  if (!(edge.ID in this.elements))
    return;

  for (var i in this.links) {
    if (this.links[i].source.ID == edge.Parent.ID &&
      this.links[i].target.ID == edge.Child.ID)
    this.links.splice(i, 1);
  }
  delete this.elements[edge.ID];

  this.Redraw();
};

Layout.prototype.Tick = function(e) {
  this.link.attr("d", this.linkArc);

  this.node.attr("cx", function(d) { return d.x; })
  .attr("cy", function(d) { return d.y; });

  this.node.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });

  var _this = this;
  if (!this.group.empty())
    this.group.data(this.Groups()).attr("d", function(d) {
      return _this.DrawCluster(d);
    });
};

Layout.prototype.linkArc = function(d) {
  var dx = d.target.x - d.source.x,
      dy = d.target.y - d.source.y,
      dr = Math.sqrt(dx * dx + dy * dy) * 1.3;
  return "M" + d.source.x + "," + d.source.y + "A" + dr + "," + dr + " 0 0,1 " + d.target.x + "," + d.target.y;
};

Layout.prototype.CircleSize = function(d) {
  switch(d.Metadata.Type) {
    case "host":
      return "22";
    case "port":
    case "ovsport":
      return 18;
    case "switch":
    case "ovsbridge":
      return 20;
    default:
      return 16;
  }
};

Layout.prototype.GroupClass = function(d) {
  return "group " + d.Type;
};

Layout.prototype.NodeClass = function(d) {
  clazz = "node " + d.Metadata.Type;

  if (d.ID in alerts)
    clazz += " alert";

  if (d.Metadata.State == "DOWN")
    clazz += " down";

  return clazz;
};

Layout.prototype.EdgeClass = function(d) {
  if (d.edge.Metadata.Type == "fabric") {
    if ((d.edge.Parent.Metadata.Probe == "fabric" && !d.edge.Child.Metadata.Probe) ||
      (!d.edge.Parent.Metadata.Probe && d.edge.Child.Metadata.Probe == "fabric")) {
        return "link local2fabric";
      }
  }

  return "link " + (d.edge.Metadata.Type || '')  + " " + (d.edge.Metadata.RelationType || '');
};

Layout.prototype.CircleOpacity = function(d) {
  if (d.Metadata.Type == "netns" && d.Metadata.Manager === null)
    return 0.0;
  return 1.0;
};

Layout.prototype.EdgeOpacity = function(d) {
  if (d.source.Metadata.Type == "netns" || d.target.Metadata.Type == "netns")
    return 0.0;
  return 1.0;
};

Layout.prototype.NodeManagerPicto = function(d) {
  switch(d.Metadata.Manager) {
    case "docker":
      return dockerImg;
    case "neutron":
      return neutronImg;
  }
};

Layout.prototype.NodeManagerStyle = function(d) {
  switch(d.Metadata.Manager) {
    case "docker":
      return "";
    case "neutron":
      return "";
  }

  return "visibility: hidden";
};

Layout.prototype.NodePicto = function(d) {
  switch(d.Metadata.Type) {
    case "host":
      return hostImg;
    case "port":
    case "ovsport":
      return portImg;
    case "bridge":
      return bridgeImg;
    case "switch":
    case "ovsbridge":
      return switchImg;
    case "netns":
      return nsImg;
    case "veth":
      return vethImg;
    case "bond":
      return portImg;
    case "container":
      return dockerImg;
    default:
      return intfImg;
  }
};

Layout.prototype.NodeProbeStatePicto = function(d) {
  if (d.IsCaptureOn())
    return probeIndicatorImg;
  return "";
};

Layout.prototype.NodeStatePicto = function(d) {
  if (d.Metadata.Type != "netns")
    return "";

  if (d.Collapsed)
    return plusImg;
  return minusImg;
};

Layout.prototype.ParentNodeForGroups = function(node) {
  var parent;
  for (var i in node.Edges) {
    var edge = node.Edges[i];
    if (edge.Parent == node)
      continue;

    if (edge.Parent.Metadata.Probe == "fabric")
      continue;

    switch (edge.Parent.Metadata.Type) {
      case "ovsport":
        break;
      case "ovsbridge":
      case "netns":
        return edge.Parent;
      default:
        parent = edge.Parent;
    }
  }

  return parent;
};

Layout.prototype.addNodeToGroup = function(parent, node, groups) {
  if (typeof parent != "undefined") {
    groupID = parent.ID;

    if (groupID in groups) {
      group = groups[groupID];
      group[node.ID] = node;

      if (node.Group === '')
        node.Group = groupID;
    }

    if (parent.Metadata.Type != "host") {
      parent = this.ParentNodeForGroups(parent);
      this.addNodeToGroup(parent, node, groups);
    }
  }
};

Layout.prototype.NodesGroups = function() {
  var groups = {};
  var node;
  var i;

  for (i in this.graph.Nodes) {
    node = this.graph.Nodes[i];

    if (!(node.ID in this.elements))
      continue;

    node.Group = '';

    var group;
    var groupID;

    // create groups and add groups owning them
    if (node.Metadata.Probe == "fabric") {
      if ("Group" in node.Metadata && node.Metadata.Group !== "") {
        groupID = node.Metadata.Group;
      } else {
        groupID = "fabric";
      }

      group = groups[groupID] || (groups[groupID] = new Group(groupID, "fabric"));
      group[node.ID] = node;
    } else {
      switch (node.Metadata.Type) {
        case "host":
        case "ovsbridge":
        case "netns":
          groupID = node.ID;
          group = groups[groupID] || (groups[groupID] = new Group(groupID, node.Metadata.Type));
          group[node.ID] = node;

          node.Group = groupID;
          break;
      }
    }
  }

  // place nodes in groups
  for (i in this.graph.Nodes) {
    node = this.graph.Nodes[i];

    if (!(node.ID in this.elements))
      continue;

    var parent = this.ParentNodeForGroups(node);
    this.addNodeToGroup(parent, node, groups);
  }

  return groups;
};

Layout.prototype.Groups = function() {
  var groupsMap = {};
  var ID;

  var groups = this.NodesGroups();
  for (ID in groups) {
    var group = groups[ID];
    for (var n in group) {
      var node = group[n];

      if (isNaN(parseFloat(node.x)))
        continue;

      if (!node.Visible)
        continue;

      // padding around group path
      var pad = 24;
      if (group.Type == "host")
        pad = 48;
      if (group.Type == "fabric")
        pad = 60;

      var g = groupsMap[ID] || (groupsMap[ID] = []);
      g.push([node.x - pad, node.y - pad]);
      g.push([node.x - pad, node.y + pad]);
      g.push([node.x + pad, node.y - pad]);
      g.push([node.x + pad, node.y + pad]);
    }
  }

  var groupsArray = [];
  for (ID in groupsMap) {
    groupsArray.push({Group: ID, Type: groups[ID].Type, path: d3.geom.hull(groupsMap[ID])});
  }

  return groupsArray;
};

Layout.prototype.DrawCluster = function(d) {
  var curve = d3.svg.line()
  .interpolate("cardinal-closed")
  .tension(0.90);

  return curve(d.path);
};

Layout.prototype.GetNodeText = function(d) {
  var name = this.graph.GetNode(d.ID).Metadata.Name;
  if (name.length > 10)
    name = name.substr(0, 8) + ".";

  return name;
};

Layout.prototype.MouseOverNode = function(d) {
  var _this = this;
  NodeDetailsTmID = setTimeout(function(){ _this.NodeDetails(d); }, 300);
};

Layout.prototype.MouseOutNode = function(d) {
  clearTimeout(NodeDetailsTmID);
};

Layout.prototype.CollapseNetNS = function(node) {
  for (var i in node.Edges) {
    var edge = node.Edges[i];

    if (edge.Child == node)
      continue;

    if (Object.keys(edge.Child.Edges).length == 1) {
      edge.Child.Visible = edge.Child.Visible ? false : true;
      edge.Visible = edge.Visible ? false : true;

      node.Collapsed = edge.Child.Visible ? false : true;
    }
  }
};

Layout.prototype.CollapseNode = function(d) {
  if (d3.event.defaultPrevented)
    return;

  switch(d.Metadata.Type) {
    case "netns":
      this.CollapseNetNS(d);
  }

  this.Redraw();
};

Layout.prototype.Redraw = function() {
  var _this = this;

  this.link = this.link.data(this.links, function(d) { return d.source.ID + "-" + d.target.ID; });
  this.link.exit().remove();

  this.link.enter().append("path")
    .style("opacity", function(d) {
      return _this.EdgeOpacity(d);
    })
    .attr("class", function(d) {
      return _this.EdgeClass(d);
    });

  this.node = this.node.data(this.nodes, function(d) { return d.ID; })
    .attr("class", function(d) {
      return _this.NodeClass(d);
    })
    .style("display", function(d) {
      return !d.Visible ? "none" : "block";
    });
  this.node.exit().remove();

  var nodeEnter = this.node.enter().append("g")
    .attr("class", "node")
    .on("click", function(d) {
      return _this.CollapseNode(d);
    })
    .on("mouseover", function(d) {
      d3.select(this).select("circle").transition()
        .duration(400)
        .attr("r", _this.CircleSize(d) * 1.2);
      _this.MouseOverNode(d);
    })
    .on("mouseout", function(d) {
      d3.select(this).select("circle").transition()
        .duration(400)
        .attr("r", _this.CircleSize(d));
      _this.MouseOutNode(d);
    })
    .call(this.drag);

  nodeEnter.append("circle")
    .attr("r", this.CircleSize)
    .attr("class", "circle")
    .style("opacity", function(d) {
      return _this.CircleOpacity(d);
    });

  nodeEnter.append("image")
    .attr("class", "picto")
    .attr("xlink:href", function(d) {
      return _this.NodePicto(d);
    })
    .attr("x", -10)
    .attr("y", -10)
    .attr("width", 20)
    .attr("height", 20);

  nodeEnter.append("image")
    .attr("class", "probe")
    .attr("x", 10)
    .attr("y", -20)
    .attr("width", 20)
    .attr("height", 20);

  nodeEnter.append("image")
    .attr("class", "state")
    .attr("x", -20)
    .attr("y", -20)
    .attr("width", 12)
    .attr("height", 12);

  nodeEnter.append("circle")
    .attr("class", "manager")
    .attr("r", 12)
    .attr("cx", 14)
    .attr("cy", 16);

  nodeEnter.append("image")
    .attr("class", "manager")
    .attr("x", 4)
    .attr("y", 6)
    .attr("width", 20)
    .attr("height", 20);

  nodeEnter.append("text")
    .attr("dx", 22)
    .attr("dy", ".35em")
    .text(function(d) {
      return _this.GetNodeText(d);
    });

  // bounding boxes for groups
  this.groupsG.selectAll("path.group").remove();
  this.group = this.groupsG.selectAll("path.group")
    .data(this.Groups())
    .enter().append("path")
    .attr("class", function(d) {
      return _this.GroupClass(d);
    })
    .attr("id", function(d) {
      return d.group;
    })
    .attr("d", function(d) {
      return _this.DrawCluster(d);
    });

  this.node.select('text')
    .text(function(d){
        return _this.GetNodeText(d);
    });

  this.node.select('image.state').attr("xlink:href", function(d) {
    return _this.NodeStatePicto(d);
  });

  this.node.select('image.probe').attr("xlink:href", function(d) {
    return _this.NodeProbeStatePicto(d);
  });

  this.node.select('image.manager').attr("xlink:href", function(d) {
    return _this.NodeManagerPicto(d);
  });

  this.node.select('circle.manager').attr("style", function(d) {
    return _this.NodeManagerStyle(d);
  });

  this.force.start();
};

Layout.prototype.ProcessGraphMessage = function(msg) {
  if (!this.live && msg.Type != "SyncReply") {
    console.log("Skipping message " + msg.Type);
    return;
  }

  var node;
  var edge;

  switch(msg.Type) {
    case "SyncReply":
      $("#node-details").hide();

      this.Clear();
      this.InitFromSyncMessage(msg);
      break;

    case "NodeUpdated":
      node = this.graph.GetNode(msg.Obj.ID);

      this.UpdateNode(node, msg.Obj.Metadata);
      break;

    case "NodeAdded":
      node = this.graph.NewNode(msg.Obj.ID, msg.Obj.Host);
      if ("Metadata" in msg.Obj)
        node.Metadata = msg.Obj.Metadata;

      this.AddNode(node);
      break;

    case "NodeDeleted":
      node = this.graph.GetNode(msg.Obj.ID);
      if (typeof node == "undefined")
        return;

      this.graph.DelNode(node);
      this.DelNode(node);

      if (typeof CurrentNodeDetails != "undefined" && CurrentNodeDetails.ID == node.ID)
        $("#node-details").hide();
      break;

    case "EdgeUpdated":
      edge = this.graph.GetEdge(msg.Obj.ID);
      edge.Metadata = msg.Obj.Metadata;

      this.Redraw();
      break;

    case "EdgeAdded":
      var parent = this.graph.GetNode(msg.Obj.Parent);
      var child = this.graph.GetNode(msg.Obj.Child);

      edge = this.graph.NewEdge(msg.Obj.ID, parent, child, msg.Obj.Host);
      if ("Metadata" in msg.Obj)
        edge.Metadata = msg.Obj.Metadata;

      this.AddEdge(edge);
      break;

    case "EdgeDeleted":
      edge = this.graph.GetEdge(msg.Obj.ID);
      if (typeof edge == "undefined")
        break;

      this.graph.DelEdge(edge);
      this.DelEdge(edge);
      break;
  }
};

Layout.prototype.ProcessAlertMessage = function(msg) {
  var _this = this;

  var ID  = msg.Obj.ReasonData.ID;
  alerts[ID] = msg.Obj;
  this.Redraw();

  setTimeout(function() { delete alerts[ID]; _this.Redraw(); }, 1000);
};

Layout.prototype.SyncRequest = function(t) {
  var msg = {"Namespace": "Graph", "Type": "SyncRequest", "Obj": {"Time": t}};
  this.updatesocket.send(JSON.stringify(msg));
};

Layout.prototype.StartLiveUpdate = function() {
  this.live = true;
  this.updatesocket = new WebSocket("ws://" + location.host + "/ws");

  var _this = this;
  this.updatesocket.onopen = function() {
    _this.SyncRequest(Date.now());
  };

  this.updatesocket.onclose = function() {
    setTimeout(function() { _this.StartLiveUpdate(); }, 1000);
  };

  this.updatesocket.onmessage = function(e) {
    var msg = jQuery.parseJSON(e.data);
    switch(msg.Namespace) {
      case "Graph":
        _this.ProcessGraphMessage(msg);
        break;
      case "Alert":
        _this.ProcessAlertMessage(msg);
        break;
    }
  };
};

var topologyLayout;
var conversationLayout;
var discoveryLayout;

function AgentReady() {
  $(".analyzer-only").hide();
}

function AnalyzerReady() {
  conversationLayout = new ConversationLayout(".conversation-d3");
  discoveryLayout = new DiscoveryLayout(".discovery-d3");

  $('#topology-btn').click(function() {
    $('#topology').addClass('active');
    $('#conversation').removeClass('active');
    $('#discovery').removeClass('active');

    $('.topology').show();
    $('.conversation').hide();
    $('.discovery').hide();
  });

  $(".title-capture-switch").hide();

  $('#conversation-btn').click(function() {
    $('#topology').removeClass('active');
    $('#conversation').addClass('active');
    $('#discovery').removeClass('active');

    $('.topology').hide();
    $('.conversation').show();
    $('.discovery').hide();

    conversationLayout.ShowConversation("ethernet");
  });
  $('#discovery-btn').click(function() {
    $('#topology').removeClass('active');
    $('#conversation').removeClass('active');
    $('#discovery').addClass('active');

    $('.topology').hide();
    $('.conversation').hide();
    $('.discovery').show();

    discoveryLayout.DrawChart();
  });
}

function Logout() {
  window.location.href = "/login";
}

function StartCheckAPIAccess() {
  setInterval(function() {
    $.ajax({
      dataType: "json",
      url: '/api',
      error: function(e) {
        if (e.status == 401)
          Logout();
      }
    });
  }, 5000);
}

var Captures = {};
function RefreshCaptureList() {
  $.ajax({
    dataType: "json",
    url: '/api/capture',
    contentType: "application/json; charset=utf-8",
    method: 'GET',
    success: function(data) {
      var clist = $('.capture-list');

      var key;
      for (key in Captures) {
        if (! (key in data)) {
          var id = "#" + key;
          $(id).remove();
          delete Captures[key];
        }
      }
      for (key in data) {
        if (!(key in Captures)) {
          Captures[key] = data[key];
          var li = $('<li/>', {id: key})
            .addClass('capture-item')
            .appendTo(clist);

          var item = $('<div/>').appendTo(li);
          var capture = $('<div/>').appendTo(item);
          var title = $('<div/>').addClass("capture-title").html(key).appendTo(capture);
          var trash = $('<div/>').addClass("capture-trash").css({"text-align": "right", "float": "right"}).appendTo(title);
          $('<div/>').addClass("capture-content").html("Gremlin Query: " + data[key].GremlinQuery).appendTo(capture);
          if (data[key].Name !== "undefined")
            $('<div/>').addClass("capture-content").html("Name: " + data[key].Name).appendTo(capture);
          if (data[key].Description !== "undefined")
            $('<div/>').addClass("capture-content").html("Description: " + data[key].Description).appendTo(capture);
          if (data[key].Type !== "undefined")
            $('<div/>').addClass("capture-content").html("Type: " + data[key].Type).appendTo(capture);

          var img = $('<img/>', {src:trashImg, width: 24, height: 24}).appendTo(trash);
          img.css('cursor', 'pointer').click(function(e) {
            var li = $(this).closest('li');
            var id = li.attr('id');

            $.ajax({
              url: '/api/capture/' + id + '/',
              contentType: "application/json; charset=utf-8",
              method: 'DELETE'
            });
            li.remove();
            delete Captures[id];
          });
        }
      }
    }
  });
}

function SetupCaptureList() {
  var resetCaptureForm = function() {
    $("#capturename").val("");
    $("#capturedesc").val("");
    $("#capturetype").val("");
    $("select#capturetype option[value != '']").remove();
  };

  var getCaptureTypes = function(type) {
    switch(type) {
      case "internal":
      case "tun":
      case "bridge":
      case "device":
      case "veth":
        return ["afpacket", "pcap"];
      case "ovsbridge":
        return ["ovssflow"];
    }
  };

  $("#cancel").click(function(e) {
    $("#capture").slideToggle(500, function () {});
    resetCaptureForm(e);
  });

  $("#add-capture").click(function(e) {
    $("#capture").slideToggle(500, function () {});

    var query = "G.V().Has('TID','" + CurrentNodeDetails.Metadata.TID + "')";
    $("#capturequery").val(query);

    var captureTypes = getCaptureTypes(CurrentNodeDetails.Metadata.Type);
    for (var t in captureTypes) {
      $("select#capturetype").append($("<option>").val(captureTypes[t]).html(captureTypes[t]));
    }
  });

  $("#create").click(function(e) {
    var name = $("#capturename").val();
    var desc = $("#capturedesc").val();
    var query = $("#capturequery").val();
    var type = $("#capturetype").val();
    if (query === "") {
      alert("Gremlin query can't be empty");
    } else {
      $.ajax({
        dataType: "json",
        url: '/api/capture',
        data: JSON.stringify({"GremlinQuery": query, "Name": name, "Description": desc, "Type": type}),
        contentType: "application/json; charset=utf-8",
        method: 'POST',
      });
      $("#capture").slideToggle(500, function () {});
      resetCaptureForm(e);
    }
  });
  setInterval(RefreshCaptureList, 1000);
}

function SetupFlowRefresh() {
  $("#flow-refresh").click(function(e) {
    ShowNodeFlows(CurrentNodeDetails);
  });
}

$(document).ready(function() {
  if (Service == "agent") {
    AgentReady();
  }
  else {
    AnalyzerReady();
  }

  $('.content').resizable({
    handles: 'e',
    minWidth: 300,
    resize:function(event,ui){
      var x=ui.element.outerWidth();
      var y=ui.element.outerHeight();
      var ele=ui.element;
      var factor = $(this).parent().width()-x;
      var f2 = $(this).parent().width() * 0.02999;
      $.each(ele.siblings(),function(idx,item) {
        ele.siblings().eq(idx).css('height',y+'px');
        ele.siblings().eq(idx).width((factor-f2)+'px');
      });
    }
  });

  $('.conversation').hide();
  $('.discovery').hide();

  topologyLayout = new Layout(".topology-d3");
  topologyLayout.StartLiveUpdate();

  StartCheckAPIAccess();

  if (Service != "agent") {
    SetupTimeSlider();
    SetupFlowRefresh();
    SetupCaptureList();
  }
});
