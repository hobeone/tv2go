angular.module('shows', [
  'shows.create',
  'tv2go.models.shows',
])
.config(function($stateProvider){
  'use strict';
  $stateProvider
    .state('tv2go.shows', {
      url: '/shows',
      views: {
        'top@tv2go' : { templateUrl: 'nav.tmpl.html',},
        'left@tv2go': {
          controller: 'ShowsListCtrl as showsListCtrl',
          templateUrl: 'shows/shows.tmpl.html',
        },
        'detail@tv2go': {
          template: '',
        },
      }
    });
})
.controller('ShowsListCtrl', function ShowsListCtrl($scope, ShowsModel){
  'use strict';
  var showsListCtrl = this;
  ShowsModel.getShows()
  .then(function(result){
    showsListCtrl.shows = result;
  });
})
;
