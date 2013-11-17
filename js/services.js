'use strict';

var photosServices = angular.module('photosServices', ['ngResource']);

photosServices.factory('Photos', ['$resource',
  function($resource){
    return $resource('/api/photos/:directory', {}, {
	    get: { method:'GET', isArray:true}}
    );
  }]);
photosServices.factory('Photo', ['$resource',
  function($resource){
    return $resource('/api/photo', {}, {
        tag: { method:'POST', params:{fname:"@fname", tag:"@tag"}, isArray: false},
        untag: { method:'POST', params:{fname:"@fname", untag:"@tag"}, isArray: false}
    });
  }]);
photosServices.factory('Edit', ['$resource',
  function($resource){
    return $resource('/api/edit', {}, {
        new: { method:'POST', params:{id: "@id", dir:"@dir", name:"@name"}, isArray: false},
    });
  }]);
