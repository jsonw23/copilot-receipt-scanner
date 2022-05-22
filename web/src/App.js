import React, { useCallback, useState } from 'react'
import { useDropzone } from 'react-dropzone'
import axios from 'axios'

import { Box, CircularProgress } from '@mui/material';


import './App.css';
import ReceiptIcon from '@mui/icons-material/Receipt';
import CheckCircleIcon from '@mui/icons-material/CheckCircle'

let conn

function ws(imageID, setStatus) {
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
  conn = new WebSocket(wsUrl)
  conn.onmessage = (msg) => {
    console.log(msg)
    let data = JSON.parse(msg.data)
    setStatus(data.status)
  }
  conn.onopen = () => {
    console.log("websocket connection opened")
  }
}


function App() {
  const [imageID, setImageID] = useState()
  const [status, setStatus] = useState()

  console.log(imageID)
  if (imageID) {
    ws(imageID, setStatus)
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

  if (!imageID) {
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
    )
  } else {
    return (
      <div className="App">
        <Box sx={{
          backgroundColor: "primary.dark",
          color: "primary.light",
          height: "100vh",
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
        }}>
          {status != "Accepted" &&
            <CircularProgress sx={{fontSize: "128px", color: "white"}} />}
          {status == "Accepted" &&
            <CheckCircleIcon sx={{fontSize: "128px", color: "white"}} />}
          {status ?? "Uploaded"}
        </Box>
      </div>
    )
  }
}

export default App;
