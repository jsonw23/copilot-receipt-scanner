import React, { useCallback, useState } from 'react'
import { useDropzone } from 'react-dropzone'
import axios from 'axios'

import { Box } from '@mui/material';


import './App.css';
import ReceiptIcon from '@mui/icons-material/Receipt';

let conn

function ws(imageID) {
  if (conn) {
    return
  }
  console.log("opening websocket connection")
  let loc = window.location, wsUrl
  if (loc.protocol === "https:") {
    wsUrl = "wss:"
  } else {
    wsUrl = "ws:"
  }
  wsUrl += "//" + loc.host
  wsUrl += `/imageStatus/${imageID}/ws`
  //wsUrl = `ws://wmcxekkb6j.us-east-1.awsapprunner.com/imageStatus/${imageID}/ws`
  conn = new WebSocket(wsUrl)
  conn.onmessage = (msg) => {
    console.log(msg)
  }
  conn.onopen = () => {
    console.log("websocket connection opened")
  }
}


function App() {
  const [imageID, setImageID] = useState()

  console.log(imageID)
  if (imageID) {
    ws(imageID)
  }


  const onDrop = useCallback((acceptedFiles) => {
    acceptedFiles.forEach((file) => {
      let formData = new FormData()
      formData.append("file", file)
      axios.post("/upload", formData, {
        headers: {
          "Content-Type": "multipart/form-data"
        },
        onUploadProgress: (progressEvent) => {
          console.log(progressEvent)
        }
      }).then(resp => {
        setImageID(resp.data.imageID)
      })
    })
  }, [])
  const { getRootProps, getInputProps } = useDropzone({ onDrop })

  return (
    <div className="App" {...getRootProps()}>
      <Box sx={{
        backgroundColor: "primary.dark",
        color: "primary.light",
        height: "100vh",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
      }}>
        <input {...getInputProps()} />
        <ReceiptIcon sx={{fontSize: "128px", color: "white"}} />
        Tap to Scan Receipt
      </Box>
    </div>
  );
}

export default App;
