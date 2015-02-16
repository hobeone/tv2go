angular.module('shows', [
  'tv2go.models.shows',
])
.config(function($stateProvider){
  $stateProvider
    .state('tv2go.shows', {
      url: '/',
      views: {
        "shows@": {
          controller: "ShowsListCtrl as showsListCtrl",
          templateUrl: "shows/shows.tmpl.html",
        },
      }
    });
})
.controller("ShowsListCtrl", function ShowsListCtrl($scope, ShowsModel){
  var ShowsListCtrl = this;
  ShowsModel.getShows()
  .then(function(result){
    ShowsListCtrl.shows = result;
  });
})
;
