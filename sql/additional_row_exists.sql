select
  true
from
  articles
where
  (!(:use_after) or id < :id_after)
  and (!(:use_before) or id > :id_before)
order by 1 %v
limit 1 offset :num_rows