--
-- Created by IntelliJ IDEA.
-- User: vvf
-- Date: 08.11.16
-- Time: 13:47
--

local function count_mapper(out, rec)
    return 1
end
local function summarizer(summ, item)
    return summ + item
end

function Count(stream)
   return stream : map(count_mapper) : aggregate(0, summarizer)
end

