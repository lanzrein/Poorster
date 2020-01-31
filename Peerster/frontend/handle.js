//Handle for the server.
//@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

////@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent

//Some regexes to verify if ip or ports are ok
const regexIP = /^(?!0)(?!.*\.$)((1?\d?\d|25[0-5]|2[0-4]\d)(\.|$)){4}$/;
const regexPort = /^[0-9]+$/;

//host to which we connect hard-coded
let host = "http://127.0.0.1:";


console.log("Handle loaded.");

function searchedfiledownload() {
  //Downloads a searched file.
  data = this.textContent;
  array = data.split(",");
  filename = array[0];
  hash = array[1].slice(2);
  console.log("Request to download file : " , filename, "h:", hash[hash.length-1], "l:",hash.length);
  tosend = {"Filename" : filename, "Request" : metahash};
  $.post(host+"/downloadfoundfile",JSON.stringify(tosend)).done(function(data) {
    //response
    console.log("got answer :" , data);
  });




}

$(document).ready(function(){
  //bind the buttons to the functions
  let port = prompt("Enter port used for GUI", "8000");
  host = host + port;
  $("#buttonNode").click(addnode);
  $("#sendmsg").click(send);

  $("li[id=private]").click(showprivate);


  $("#sendprivate").click(sendprivatemsg);

  $("#cancelprivate").click(cancelprivate);

  $("#upload_btn").click(UploadFile);
  $("#downloadButton").click(DownloadFile);

  //hw3

  $("li[id=file]").dblclick(searchedfiledownload);
  $("#search_btn").click(SearchFile);


  //project

  $("#clusterbutton").click(showclusterpannel);
  $("#closecluster").click(closeclusterpannel);

  $("#initcluster").click(InitCluster);
  $("#leavecluster").click(LeaveCluster);
  $("#sendbroadcast").click(SendBroadcast);

  $("li span[class=anoncall]").click(anoncall);
  $("li span[class=anonmessage]").click(anonmessage);
  $("li span[class=expellmember]").click(expellmember);

  $("#acceptvote").click(castvote);
  $("#denyvote").click(castvote);


  $("li[id=vote]").click(openvotepannel);
  $("#closevote").click(closevote);
  $("#joincluster").click(openjoin);
  $("#confirmjoin").click(joinrequest);
  $("#closejoin").click(closejoin);


  //For calls
  $("#decline_call").click(decline_call);
  $("#accept_call").click(accept_call);
  $("#hangup_call").click(hangup_call);
  $("#dial_call").click(dial_call);

  //request the node ID
  requestNodeId();
  // //loop and request data sometimes.
  loop();




});

function SearchFile(){
  console.log("Send search request")
  jqueryBudget = $("#budget");
  jquerykeywords = $("#keywords");
  budget = jqueryBudget.val();
  keywords = jquerykeywords.val();

  console.log("Keywords :" + keywords);
  console.log("Budget" + budget);
  if (keywords === "" || budget < 0 ){
    alert("Keywords should not be empty and budget should be >= 0")

  }else{
    //send the request
    tosend = {"Keywords" : keywords, "Budget" : Number(budget)};
    $.post(host+"/searchfile",JSON.stringify(tosend)).done(function(data) {
      //update the peer list...
      console.log("OK for search");
    });

  }



  //clear the values
  jqueryBudget.val(0);
  jquerykeywords.val("");
}

/**
 *
 * Uploads a file to be indexed by the gossiper
 */
function UploadFile(){
  try{
    filename = $("#file")[0].files[0].name;
    console.log(filename);
    if (filename == ""){
      return;
    }

    $.post(host+"/sharefile",filename).done(function(data){
      //response from the upload
      if (data == "OK"){
        console.log("File uploaded");
      }else{
        console.log("Error");
      }

    });
    $("#file")[0].files[0].name = "";
  }catch{
    console.log("No file selected")
  }

}

/**
 * Request a file to be downloaded
 * @constructor
 */
function DownloadFile(){
  let filename = $("#filename").val();
  let destination = $("#destination").val();
  let request = $("#request").val();
  console.log("DL file "+filename+" from "+destination+" req "+request);

  //Send the request.
  tosend = {"Filename" : filename, "Destination" : destination, "Request" : request}
  $.post(host+"/downloadfile",JSON.stringify(tosend)).done(function(data){
    //update the peer list...
    console.log("File download got response  : " + data)


  });




  //reseting the values
  $("#filename").val('');
  $("#destination").val('');
  $("#request").val('');
}

/**
 * Cancel sending a private message
 */
function cancelprivate(){
  if (anonFlag){
    anonFlag = false;
    $("#clusterpopup").show();
    $("#anonparams").hide();

  }
  $("#receiver").text('');
  $("#privatecontent").val('');
  $("#privatepopup").hide();
}

/**
 * Send a private message from the client
 */
function sendprivatemsg(){
  let destination = $("#receiver").text();
  let content = $("#privatecontent").val();
  console.log("Sent " + content + " to : " + destination);
  tosend = {"destination" : destination, "content" : content}

  if (!anonFlag) {
    //Private message.
    $.post(host+"/privatemsg",JSON.stringify(tosend)).done(function(data){
      //update the peer list...
      console.log("Private Message got response : " + data)
      $("#origin_list").empty();
      res = JSON.parse(data);

      for(let i = 0 ; i < res.length ; i ++){

        $("#origin_list").append("<li id='private'>"+res[i]+"</li>");

      }
      $("li[id=private]").click(showprivate);


    });
  }else{
    fullAnon = $("#fullAnon").is(':checked') ? true : false ;
    relayRate = $("#relayrate").val();
    console.log("Fullanon"+fullAnon);
    console.log("relay rate :" + relayRate);
    tosend = {"destination" : destination, "content" : content, "RelayRate" : Number(relayRate), "fullAnon":fullAnon}
    console.log("To send : " + JSON.stringify(tosend))
    $.post(host+"/anonmessage",JSON.stringify(tosend)).done(function(data){
      //update the peer list...
      console.log("Anonymous Message got response : " + data)
      $("#clustermembers").empty();
      res = JSON.parse(data);

      for(let i = 0 ; i < res.length ; i ++){
        if (res[i] === currently_calling){
          $("#clustermembers").append("<li id='member' style='color: crimson'>"+res[i]+"<span class=\"anonmessage\"> ✉️</span><span class=\"anoncall\">📞</span><span class=\"expellmember\">❌</span></li>");

        }else{
          $("#clustermembers").append("<li id='member'>"+res[i]+"<span class=\"anonmessage\"> ✉️</span><span class=\"anoncall\">📞</span><span class=\"expellmember\">❌</span></li>");

        }

      }

      $("li span[class=anoncall]").click(anoncall);
      $("li span[class=anonmessage]").click(anonmessage);
      $("li span[class=expellmember]").click(expellmember);


    });
    $("#clusterpopup").show();
    $("#anonparams").hide();

  }

  //closing
  $("#privatecontent").val(" ");
  $("#receiver").text('');

  $("#privatepopup").hide();
  anonFlag = false;
}

/**
 * Show the private popup
 */
function showprivate(){
  destination = $(this).text();
  console.log(destination);
  $("#receiver").text(destination);
  $("#privatepopup").show();

}

function showclusterpannel(){


  $.get()

  $("#clusterpopup").show();



}

function closeclusterpannel(){
  $("#clusterpopup").hide();
}

/**
 * Request the node id
 */
function requestNodeId(){
  $.post(host+"/id").done(function(data){
    //update the peer list...
    res = JSON.parse(data);
    console.log("Node id is  : " + data);
    $("#nodeid").text(res);

  });

}

/**
 * Loop infinitly to get perdiodic updates from the Peeerster.
 */
function loop(){
  //request for nodes information
  $.get(host+"/node").done(function(data){
    //update the peer list...
    console.log("Node update got response : " + data);
    $("#node_list").empty();
    res = JSON.parse(data);
    for(let i = 0 ; i < res.length ; i ++){

      $("#node_list").append("<li>"+res[i]+"</li>");
    }

  });

  //query message information
  $.ajax({
    type : 'GET',
    url : host+'/message',
    success : function(data){
      console.log("update got : " + data);
      parseData(data);
    },
    complete : function(data){
      setTimeout(loop,3000);
    }
  });

  //query origins.
  $.get(host+"/privatemsg").done(function(data){
    //update the peer list...
    console.log("Private Message got response : " + data)
    $("#origin_list").empty();
    res = JSON.parse(data);


    for(let i = 0 ; i < res.length ; i ++){

      $("#origin_list").append("<li id='private'>"+res[i]+"</li>");

    }
    $("li[id=private]").click(showprivate);


  });

  //*query found files..
  $.get(host+"/searchfile").done(function(data){
    console.log("Got data :" , data)
    $("#found_files").empty();
    let res = JSON.parse(data);

    for (let i = 0 ; i < res.length; i ++){
      console.log("res : " + res[i]);
      metabytes = (res[i].MetaHash);
      x = res[i].MetaHash;
      metahash = toHexString(metabytes);
      filename = res[i].Filename;
      console.log("name : " +filename + ", h : ", metahash);
      $("#found_files").append("<li id='file'>"+filename + ",h:"+metahash+"</li>");
    }

    $("li[id=file]").dblclick(searchedfiledownload);
  });


  $.get(host+"/anonmessage").done(function(data){
    //update the peer list...
    console.log("Anonymous Message got response : " + data)
    $("#clustermembers").empty();
    res = JSON.parse(data);

    for(let i = 0 ; i < res.length ; i ++){
      if (res[i] === currently_calling){
        $("#clustermembers").append("<li id='member' style='color: crimson'>"+res[i]+"<span class=\"anonmessage\"> ✉️</span><span class=\"anoncall\">📞</span><span class=\"expellmember\">❌</span></li>");

      }else{
        $("#clustermembers").append("<li id='member'>"+res[i]+"<span class=\"anonmessage\"> ✉️</span><span class=\"anoncall\">📞</span><span class=\"expellmember\">❌</span></li>");

      }
    }

    $("li span[class=anoncall]").click(anoncall);
    $("li span[class=anonmessage]").click(anonmessage);
    $("li span[class=expellmember]").click(expellmember);


  });

  $.get(host+"/evoting").done(function(data){
    //update the peer list...
    console.log("Evoting got response : " + data)
    $("#votelist").empty();
    res = JSON.parse(data);

    for(let i = 0 ; i < res.length ; i ++){

      $("#votelist").append("<li id='vote'>"+res[i]+"</li>");

    }
    $("li[id=vote]").click(openvotepannel);

  });



  $.get(host+"/incomingcall").done(function(data){
    //update the peer list...
    console.log("Calling got response : " + data)
    res = JSON.parse(data)
    if (res !== "" && currently_calling !== res){
      $("#callpannel").show();
      $("#callee").text(data);
      currently_calling = data;
    }

  });

}


var x;

/**
 * Parse the data and append it to the messages
 * @param data to parse . expected to be \n separeated lines
 */
function parseData(data){
  lines = data.split("\n");

  for(let i=0; i < lines.length; i++){
      //its a message.
      toadd = "<p>"+lines[i]+"</p>";
      console.log("adding : " + toadd );
      $(toadd).appendTo('#messages');
    }
}

/**
 * Add a new node
 */
function addnode(){
  console.log("adding node");
  node = $("#nodevalue").val();
  console.log("value is : " +node );
  //check input...

  if (node == "" || node == "ip:port") {
    console.log("Wrong node value");
    display_node_message(node);
    return ;
  }

  pairs = node.split(":")
  if (pairs.length != 2){
    display_node_message(node);
    return ;
  }

  ip = pairs[0];
  port = pairs[1];

  if (regexIP.test(ip) && regexPort.test(port)){
    //we have a valid input.. POST IT
    $.post(host+"/node",node).done(function(data){
      //update the peer list...
      console.log("Node adding got response : " + data)
      $("#node_list").empty();
      res = JSON.parse(data)
      console.log(" RES = " + res)
      for(let i = 0 ; i < res.length ; i ++){

        $("#node_list").append("<li>"+res[i]+"</li>");
      }
    });
  }


}

/**
 * Send a message to the peerster
 */
function send(){
  //get the value to send.
  console.log("Sending message " );
  msg = $("#message").val();
  console.log("Message value : " + msg);
  if (msg == ""){
    //dont do anything on empty message
    return
  }

  addr = host + "/message";
  console.log("Posting at " + addr);
  posting = $.post(addr,msg);
  //post and get the reply.
  posting.done(function(data){
    parseData(data);
  });
  console.log("Posted");
  $("#message").val("");

}



function display_node_message(str){
  alert("Expected format input : <ip>:<port> got : " + str)
}


function toHexString(bytes) {
  let string = '';
  for (let i = 0; i < 32; i++) {
    hex = (atob(bytes).charCodeAt(i) & 0xFF).toString(16);
    if (hex.length == 1){
      hex = '0' + hex;
    }
    string += hex
  }
  return string

}
