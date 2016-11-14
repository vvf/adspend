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


operators = {}
function operators.eq(a,b)
    return a==b
end
function operators.neq(a,b)
    return not a==b
end

function FilterEq(stream, fields, values)
    return stream : filter(map(), function(item)
        local result = True
        for i, field in pairs(fields) do
            result = result and (item[field] == values[i])
        end
        return result
    end)
end

function ValuesOf(stream, field)
    local function sum(a,b)
        return a+ b
    end
    return stream : aggregate(map(), function(result, item)
        local value = item[field]
        result[value] = (result[value] or 0) + 1
        return result
    end) : reduce(function(c1,c2)
        return map.merge(c1, c2, sum)
    end)
end