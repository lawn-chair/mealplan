
export const updatePosition = (arr, oPos, nPos, offset = 0) => {
    if (typeof nPos === "number" && typeof oPos === "number" && nPos !== oPos) {
      arr.splice(nPos + offset, 0, arr.splice(oPos + offset, 1)[0])
    }
  
    arr.forEach((item, index) => {
      item.order = index + 1
    })
  
    console.log("Positioned: ", arr);
  
    return arr
  }