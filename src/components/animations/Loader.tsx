import React from 'react'

function Loader() {
  return (
    <>
    <div className="flex items-center justify-center h-screen bg-black">
      <div className="flex flex-col items-center space-y-4">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <p className="text-lg text-gray-300">Loading...</p>
      </div>
    </div>
    </>
  )
}

export default Loader