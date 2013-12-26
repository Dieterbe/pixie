'use strict';

var photosServices = angular.module('photosServices', ['ngResource']);

photosServices.factory('Binds', ['$resource',
  function($resource){
    return $resource('/api/config/binds', {}, {
        list: { method:'GET', isArray: false}}
    );
  }
]);
photosServices.factory('Photos', ['$resource',
  function($resource){
    return $resource('/api/photos/:directory', {}, {
	    get: { method:'GET', isArray:true}}
    );
  }
]);
photosServices.factory('Photo', ['$resource',
  function($resource){
    return $resource('/api/photo', {}, {
        tag: { method:'POST', params:{dir:"@dir", name:"@name", tag:"@tag"}, isArray: false},
        untag: { method:'POST', params:{dir:"@dir", name:"@name", untag:"@tag"}, isArray: false}
    });
  }
]);
photosServices.factory('Edit', ['$resource',
  function($resource){
    return $resource('/api/edit', {}, {
        new: { method:'POST', params:{id: "@id", dir:"@dir", name:"@name"}, isArray: false},
    });
  }
]);
