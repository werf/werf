module Jekyll
   module Drops
     class BreadcrumbItem < Liquid::Drop
       extend Forwardable
 
       def initialize(side)
         @side = side
       end
 
       def position
         @side[:position]
       end
 
       def title
         @side[:title]
       end
 
       def url
         @side[:url]
       end

       def rootimage
         @side[:root_image]
       end
      
     end
   end
 end
 