
import {parse, format} from 'date-fns';

export function formatDate(date: string) {
  return format(parse(date, "yyyy-MM-dd", new Date), "MMMM d, yyyy");
}

export const updatePosition = (arr: {id?: number, text: string, order: number}[], oPos: number, nPos: number, offset = 0) => {
    if (typeof nPos === "number" && typeof oPos === "number" && nPos !== oPos) {
      arr.splice(nPos + offset, 0, arr.splice(oPos + offset, 1)[0])
    }
  
    arr.forEach((item, index) => {
      item.order = index + 1
    })
  
    console.log("Positioned: ", arr);
  
    return arr
  }