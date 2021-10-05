select
  true
from
  articles
where
  (!(:use_after) or id >= :id_after)
  and (!(:use_before) or id <= :id_before)
limit 1