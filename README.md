### _kew_ - K, Enough of Werc

_kew_ is an extremely minimal static site generator written in go inspired by the legendary _werc_ by Uriel @ cat-v.  
it uses `lowdown`.


**why make this when _werc_ exists?**  
_werc_ is a fantastic program but has its own tradeoffs:
- it is a dynamic web system, not a static site generator
- it requires Plan9 utils (rc-shell)
- it is slow

_kew_ fixes these by adopting a different philosophy:  
_werc_ treats the site as a dynamic system,  
_kew_ treats the site as a build output.  
