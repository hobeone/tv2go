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
        'episodes@': {
          templateUrl: 'shows/episodes/episodes.tmpl.html',
          controller: 'EpisodesListCtrl as episodesListCtrl',
        }
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
