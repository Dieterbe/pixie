'use strict';

/* Controllers */

var photosControllers = angular.module('photosControllers', []);

photosControllers.controller('PhotosCtrl', ['$scope', '$routeParams', 'Photos', 'Photo', 'Edit',
  function($scope, $routeParams, Photos, Photo, Edit) {
    $scope.setDirectory = function() {
        $scope.photos = Photos.get({directory: $scope.directory}, function(response) {
            $scope.logs.push({msg: "loaded " + $scope.directory, type: 'info'});
        }, function(response) {
            $scope.logs.push({msg: "failed to load " + $scope.directory, type: 'error'});
        });
    }
    // routeParams.dir
    $scope.focusIndex = 0; // determines position top-bottom
    $scope.subFocusIndex = 0; // determines position left (original) to right (any edits)
    $scope.logs = [];
    $scope.openRecord = function () {
        $scope.$apply(function () {
        console.log('opening : ', $scope.photos[$scope.focusIndex] );
        });
    };
    $scope.newEdit = function () {
        $scope.$apply(function () {
            console.log('creating edit for :', $scope.photos[$scope.focusIndex] );
            $scope.photos[$scope.focusIndex] = Edit.new($scope.photos[$scope.focusIndex], function(response) {
                $scope.logs.push({msg: "created new edit for " + $scope.photos[$scope.focusIndex].name, type: 'info'});
            }, function(response) {
                $scope.logs.push({msg: "failed to edit " + $scope.photos[$scope.focusIndex].name, type: 'error'});
            });
        });
    };
    $scope.moveHome = function () {
        $scope.$apply(function () {
            $scope.focusIndex = 0;
            $scope.subFocusIndex = 0;
            window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
        });
    }
    $scope.moveEnd = function () {
        $scope.$apply(function () {
            $scope.focusIndex = $scope.photos.length -1
            $scope.subFocusIndex = 0;
            window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
        });
    }
    $scope.moveDown = function () {
        $scope.$apply(function () {
            if ($scope.focusIndex < $scope.photos.length -1) {
                $scope.focusIndex++;
                $scope.subFocusIndex = 0;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
        });
    }
    $scope.moveUp = function () {
        $scope.$apply(function () {
            if($scope.focusIndex > 0 ) {
                $scope.focusIndex--;
                $scope.subFocusIndex = 0;
                window.scrollTo(0, $("#photo-" + $scope.focusIndex).offset().top - 200);
            }
        });
    }
    $scope.moveLeft = function () {
        $scope.$apply(function () {
            if($scope.subFocusIndex > 0 ) {
                $scope.subFocusIndex--;
            }
        });
    }
    $scope.moveRight = function () {
        $scope.$apply(function () {
            // 3 edits means focusIndex can be max 3 (0 original, 1/2/3 for the edits)
            if($scope.subFocusIndex < Object.keys($scope.photos[$scope.focusIndex].edits).length) {
                $scope.subFocusIndex++;
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
                $scope.logs.push({msg: response.msg + ": "  + fname + " (" + tag + ")", type: 'error'});
            });
        });
    }
    $scope.unTag = function (tag) {
        $scope.$apply(function () {
            var fname = $scope.directory + "/" + $scope.photos[$scope.focusIndex].name;
            var index = $scope.focusIndex; // not sure if needed, but by the time the callback fires we may have focused on other image
            Photo.untag({fname: fname, tag: tag}, function(response) {
                console.debug(response);
                $scope.logs.push({msg: response.msg + ": " + fname + " (" + tag + ")", type: 'info'});
                var tag_index = $scope.photos[index]['tags'].indexOf(tag)
                if(tag_index!=-1){
                       $scope.photos[index]['tags'].splice(tag_index, 1);
                }
            }, function(response) {
                console.debug(response);
                $scope.logs.push({msg: response.msg + ": " + fname + " (" + tag + ")", type: 'error'});
            });
        });
    }
    // todo $scope.autotag = function (tag) {
  }]);
photosControllers.controller('PhotoDetailCtrl', ['$scope', '$routeParams', 'Photos', 'Photo',
  function($scope, $routeParams, Photos, Photo) {
  }]);
