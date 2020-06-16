/*
 *  Copyright (c) 2015 The WebRTC project authors. All Rights Reserved.
 *
 *  Use of this source code is governed by a BSD-style license
 *  that can be found in the LICENSE file in the root of the source
 *  tree.
 */

'use strict';

// Put variables in global scope to make them available to the browser console.
// const audio = document.querySelector('audio');

// const constraints = window.constraints = {
//   audio: true,
//   video: false
// };

// function handleSuccess(stream) {
//   const audioTracks = stream.getAudioTracks();
//   console.log('Got stream with constraints:', constraints);
//   console.log('Using audio device: ' + audioTracks[0].label);
//   stream.oninactive = function() {
//     console.log('Stream ended');
//     };
//   window.stream = stream; // make variable available to browser console
//   audio.srcObject = stream;
// }

// function handleError(error) {
//   const errorMessage = 'navigator.MediaDevices.getUserMedia error: ' + error.message + ' ' + error.name;
//   var errorMsgElement = document.getElementById("errorMsg");
//   errorMsgElement.innerHTML = errorMessage;
//   console.log(errorMessage);
// }

// navigator.mediaDevices.getUserMedia(constraints).then(handleSuccess).catch(handleError);


// peerConnection.setRemoteDescription(new RTCSessionDescription(message));
// const answer = await peerConnection.createAnswer();
// await peerConnection.setLocalDescription(answer);

//This should wait for track stream and then we should embed

// var audioElem = document.querySelector('audio');
// peerConnectionpc.ontrack = ev => {
//     audioElem.srcObject = ev.streams[0];
// }


function handleSDPSubmit() {
  console.log("Submit pressed");
  var sdpelem = document.getElementById("sdp");
  var sdp = sdpelem.value;
  console.log(sdp);


  
}