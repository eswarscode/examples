append on newslice can reflect on original source array
even if have slice of one element referencing  to  orgSlice of many elements, memory won't be freed
In map if you add and delete many keys , buckets will increase and won't shrink back 