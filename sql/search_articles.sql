select
  id
from
  (
    select
      id
    from
      articles
    where
      (!(:use_after) or id < :id_after)
      and (!(:use_before) or id > :id_before)
    order by 1 %v -- lastが指定されたときのみasc, それ以外はdesc
    limit :num_rows
  ) as t
order by 1 desc