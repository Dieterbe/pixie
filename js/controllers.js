'use strict';

/* Controllers */

var photosControllers = angular.module('photosControllers', []);

photosControllers.controller('PhotosCtrl', ['$scope', '$routeParams', 'Photos', 'Photo',
  function($scope, $routeParams, Photos, Photo) {
    $scope.setDirectory = function() {
        $scope.photos = Photos.get({directory: $scope.directory}, function(response) {
            $scope.logs.push({msg: response.msg, type: 'info'});
        }, function(response) {
            $scope.logs.push({msg: response.msg, type: 'error'});
		});
    }
    // routeParams.dir
    $scope.focusIndex = 0;
    $scope.logs = [];
    $scope.openRecord = function () {
	    $scope.$apply(function () {
        console.log('opening : ', $scope.photos[$scope.focusIndex] );
	    });
    };
    $scope.moveDown = function () {
	    $scope.$apply(function () {
            if ($scope.focusIndex < $scope.photos.length -1) {
                $scope.focusIndex++;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
	    });
    }
    $scope.moveUp = function () {
	    $scope.$apply(function () {
            if($scope.focusIndex > 0 ) {
                $scope.focusIndex--;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
	    });
    }
    $scope.tag = function (tag) {
	    $scope.$apply(function () {
            var fname = $scope.directory + "/" + $scope.photos[$scope.focusIndex].name;
            var index = $scope.focusIndex; // not sure if needed, but by the time the callback fires we may have focused on other image
            Photo.tag({fname: fname, tag: tag}, function(response) {
                $scope.logs.push({msg: response.msg + ": " + fname + " (" + tag + ")", type: 'info'});
                if(response.msg != "tag already existed") {
                    $scope.photos[index]['tags'].push(tag);
                }
            }, function(response) {
                console.debug(response);
                $scope.logs.push({'msg': response.msg + ": "  + fname + " (" + tag + ")", type: 'error'});
            });
	    });
    }
    $scope.unTag = function (tag) {
	    $scope.$apply(function () {
            var fname = $scope.directory + "/" + $scope.photos[$scope.focusIndex].name;
            var index = $scope.focusIndex; // not sure if needed, but by the time the callback fires we may have focused on other image
            Photo.untag({fname: fname, tag: tag}, function(response) {
                $scope.logs.push({msg: response.msg + ": " + fname + " (" + tag + ")", type: 'info'});
                // todo$scope.photos[index]['tags'].push(tag);
            }, function(response) {
                console.debug(response);
                $scope.logs.push({'msg': response.msg + ": " + fname + " (" + tag + ")", type: 'error'});
            });
	    });
    }
    // todo $scope.autotag = function (tag) {
  }]);
