/**
 * @file cluster.js contains the handler to connect to the server and make requests for the cluster management
 //@authors Hrusanov Aleksandar, Lanzrein Johan, Rinaldi Vincent
 */

function InitCluster(){

    console.log(host);
    tosend = {Req: 0};
        //send the request
    $.post(host+"/initcluster",JSON.stringify(tosend)).done(function(data) {
        console.log("OK for init");
    });


}


function LeaveCluster(){
    console.log(host);
    tosend = {Req: 0};
    //send the request
    $.post(host+"/leavecluster",JSON.stringify(tosend)).done(function(data) {
        console.log("OK for leave");
    });
}

function SendBroadcast(){
    let content = $("#broadcastcontent").val();
    let destination = ""
    console.log("Sent " + content + " as broadcast");
    tosend = {"destination" : destination, "content" : content};
    $.post(host+"/broadcastmsg",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Broadcast message got response : " + data)
        $("#clustermembers").empty();
        res = JSON.parse(data);

        for(let i = 0 ; i < res.length ; i ++){

            $("#clustermembers").append("<li id='member'>"+res[i]+"</li>");

        }
        $("li[id=member]").click(showanonymous);


    });
    //closing
    $("#broadcastcontent").val(" ");

    $("#clusterpopup").hide();


}


let anonFlag = false;

function anonmessage(){
    dst = $(this).parent().text();
    if (dst.length > 5) {
        dst = dst.substring(0,dst.length-5);
    }
    console.log("anon message ! "+ dst ) ;

    $("#receiver").text(dst);
    $("#privatepopup").show();
    $("#clusterpopup").hide();
    $("#anonparams").show();
    anonFlag = true ;
}

let dst;
function anoncall(){
    dst = $(this).parent().text() ;
    if (dst.length > 5) {
        dst = dst.substring(0,dst.length-5);
    }
    console.log("anon call to : " + dst   );
    tosend = {"destination":dst, "content":content};
    $.post(host+"/anoncall",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Anonymous call got response : " + data)

    });

}

let person;
let vote;

function openvotepannel(){
    let text = $(this).text();
    let xs = text.split(" ");
    if (xs.length !== 2){
        console.log("Unexpected length of array");
        return ;
    }

    vote = xs[0];
    person = xs[1];

    console.log("Vote "+vote + "person "+person);
    $("#votepannel").show();
    $("#votetext").text("Do you want to "+vote+" person " + person +" to the cluster ?") ;



}
function castvote(){
    let decision = $(this).text();
    let bool = decision === 'YES' ;
    tosend = {'Vote' : vote, 'Person' : person , 'Decision' : bool}
    console.log("Sending " + JSON.stringify(tosend));

    $.post(host+"/evoting",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Evoting got response : " + data)
        $("#votelist").empty();
        res = JSON.parse(data);

        for(let i = 0 ; i < res.length ; i ++){

            $("#votelist").append("<li id='vote'>"+res[i]+"</li>");

        }
        $("li[id=vote]").click(openvotepannel);

    });

    closevote()
}

function closevote(){
    $("#votepannel").hide();
    $("#votetext").text('');
    person = '';
    vote = '';


}

function openjoin(){
    $("#joinpannel").show();

}

function joinrequest(){
    let other = $("#joinOther").val();
    if (other==""){
        alert("Empty member");
    }


    tosend = {"joinOther":other};
    $.post(host+"/joinrequest",JSON.stringify(tosend)).done(function(data){
        //update the peer list...
        console.log("Anonymous call got response : " + data)

    });

}

function closejoin(){
    $("#joinpannel").hide();
    $("#joinOther").text("");
}

$("#joincluster").click(openjoin);
$("#confirmjoin").click(joinrequest);
$("#closejoin").click(closejoin);